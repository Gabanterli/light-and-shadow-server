package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	minCharacterPositionRelocationCoordinate = -1 << 31
	maxCharacterPositionRelocationCoordinate = 1<<31 - 1

	// PostgreSQL INT is signed 32-bit. The expected version must leave room
	// for the atomic increment executed by this operation.
	maxCharacterPositionRelocationExpectedVersion = 1<<31 - 2

	characterPositionRelocationTimeout = 5 * time.Second
)

const characterPositionRelocationUpdateQuery = `
UPDATE characters
SET
    posX = $1,
    posY = $2,
    posZ = $3,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE name = $4
  AND version = $5
`

// CharacterPositionRelocationRequest describes a persistence-only correction
// whose destination was already resolved by authoritative world rules.
type CharacterPositionRelocationRequest struct {
	PlayerID        string
	X               float64
	Y               float64
	Z               float64
	ExpectedVersion int
}

// CharacterPositionRelocationResult contains the version produced by a
// successful optimistic-lock update.
type CharacterPositionRelocationResult struct {
	NewVersion int
}

// CharacterPositionRelocationValidationError reports an invalid request before
// any persistence operation begins.
type CharacterPositionRelocationValidationError struct {
	Field  string
	Reason string
}

func (e *CharacterPositionRelocationValidationError) Error() string {
	return fmt.Sprintf(
		"invalid character position relocation field %q: %s",
		e.Field,
		e.Reason,
	)
}

// CharacterPositionRelocationPersistenceUnavailableError reports that the
// durability guarantee cannot be provided because PostgreSQL is unavailable.
type CharacterPositionRelocationPersistenceUnavailableError struct{}

func (e *CharacterPositionRelocationPersistenceUnavailableError) Error() string {
	return "character position relocation requires PostgreSQL persistence"
}

// CharacterPositionRelocationConflictError represents a missing character or
// optimistic-lock version mismatch.
type CharacterPositionRelocationConflictError struct {
	PlayerID        string
	ExpectedVersion int
}

func (e *CharacterPositionRelocationConflictError) Error() string {
	return fmt.Sprintf(
		"character position relocation conflict for player %q at expected version %d",
		e.PlayerID,
		e.ExpectedVersion,
	)
}

// CharacterPositionRelocationIntegrityError reports an impossible multi-row
// update for the unique character-name boundary.
type CharacterPositionRelocationIntegrityError struct {
	PlayerID     string
	RowsAffected int64
}

func (e *CharacterPositionRelocationIntegrityError) Error() string {
	return fmt.Sprintf(
		"character position relocation updated %d rows for unique player %q",
		e.RowsAffected,
		e.PlayerID,
	)
}

type characterPositionRelocationDatabase interface {
	BeginTx(
		ctx context.Context,
		options *sql.TxOptions,
	) (characterPositionRelocationTransaction, error)
}

type characterPositionRelocationTransaction interface {
	ExecContext(
		ctx context.Context,
		query string,
		args ...any,
	) (sql.Result, error)

	Commit() error
	Rollback() error
}

type sqlCharacterPositionRelocationDatabase struct {
	database *sql.DB
}

func (d *sqlCharacterPositionRelocationDatabase) BeginTx(
	ctx context.Context,
	options *sql.TxOptions,
) (characterPositionRelocationTransaction, error) {
	return d.database.BeginTx(ctx, options)
}

// RelocateCharacterPosition persists only authoritative position and version.
// It intentionally does not rewrite inventory, combat statistics, currency,
// progression, class, affinities, quests, or dialogue state.
func (pm *PersistenceManager) RelocateCharacterPosition(
	ctx context.Context,
	request CharacterPositionRelocationRequest,
) (CharacterPositionRelocationResult, error) {
	normalizedRequest, err :=
		validateCharacterPositionRelocationRequest(
			request,
		)
	if err != nil {
		return CharacterPositionRelocationResult{}, err
	}

	if pm == nil ||
		pm.pgPool == nil ||
		pm.pgPool.DB == nil {
		return CharacterPositionRelocationResult{},
			&CharacterPositionRelocationPersistenceUnavailableError{}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	operationContext, cancel := context.WithTimeout(
		ctx,
		characterPositionRelocationTimeout,
	)
	defer cancel()

	return executeCharacterPositionRelocation(
		operationContext,
		&sqlCharacterPositionRelocationDatabase{
			database: pm.pgPool.DB,
		},
		normalizedRequest,
	)
}

func executeCharacterPositionRelocation(
	ctx context.Context,
	database characterPositionRelocationDatabase,
	request CharacterPositionRelocationRequest,
) (CharacterPositionRelocationResult, error) {
	if database == nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"character position relocation database cannot be nil",
			)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"character position relocation context is unavailable: %w",
				err,
			)
	}

	transaction, err := database.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
		},
	)
	if err != nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"begin character position relocation transaction: %w",
				err,
			)
	}

	if transaction == nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"begin character position relocation transaction returned nil transaction",
			)
	}

	committed := false

	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	result, err := transaction.ExecContext(
		ctx,
		characterPositionRelocationUpdateQuery,
		request.X,
		request.Y,
		request.Z,
		request.PlayerID,
		request.ExpectedVersion,
	)
	if err != nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"update character position for player %q: %w",
				request.PlayerID,
				err,
			)
	}

	if result == nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"update character position for player %q returned a nil SQL result",
				request.PlayerID,
			)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"read character position relocation rows for player %q: %w",
				request.PlayerID,
				err,
			)
	}

	switch rowsAffected {
	case 0:
		return CharacterPositionRelocationResult{},
			&CharacterPositionRelocationConflictError{
				PlayerID:        request.PlayerID,
				ExpectedVersion: request.ExpectedVersion,
			}

	case 1:
		// Expected unique optimistic-lock update.

	default:
		return CharacterPositionRelocationResult{},
			&CharacterPositionRelocationIntegrityError{
				PlayerID:     request.PlayerID,
				RowsAffected: rowsAffected,
			}
	}

	if err := transaction.Commit(); err != nil {
		return CharacterPositionRelocationResult{},
			fmt.Errorf(
				"commit character position relocation for player %q: %w",
				request.PlayerID,
				err,
			)
	}

	committed = true

	return CharacterPositionRelocationResult{
		NewVersion: request.ExpectedVersion + 1,
	}, nil
}

func validateCharacterPositionRelocationRequest(
	request CharacterPositionRelocationRequest,
) (CharacterPositionRelocationRequest, error) {
	request.PlayerID = strings.TrimSpace(
		request.PlayerID,
	)

	if request.PlayerID == "" {
		return CharacterPositionRelocationRequest{},
			&CharacterPositionRelocationValidationError{
				Field:  "player_id",
				Reason: "cannot be empty",
			}
	}

	var err error

	request.X, err =
		validateCharacterPositionRelocationCoordinate(
			request.X,
			"x",
		)
	if err != nil {
		return CharacterPositionRelocationRequest{}, err
	}

	request.Y, err =
		validateCharacterPositionRelocationCoordinate(
			request.Y,
			"y",
		)
	if err != nil {
		return CharacterPositionRelocationRequest{}, err
	}

	request.Z, err =
		validateCharacterPositionRelocationCoordinate(
			request.Z,
			"z",
		)
	if err != nil {
		return CharacterPositionRelocationRequest{}, err
	}

	if request.ExpectedVersion < 1 {
		return CharacterPositionRelocationRequest{},
			&CharacterPositionRelocationValidationError{
				Field: "expected_version",
				Reason: fmt.Sprintf(
					"must be at least 1, got %d",
					request.ExpectedVersion,
				),
			}
	}

	if request.ExpectedVersion >
		maxCharacterPositionRelocationExpectedVersion {
		return CharacterPositionRelocationRequest{},
			&CharacterPositionRelocationValidationError{
				Field: "expected_version",
				Reason: fmt.Sprintf(
					"must be at most %d so the PostgreSQL INT increment remains valid",
					maxCharacterPositionRelocationExpectedVersion,
				),
			}
	}

	return request, nil
}

func validateCharacterPositionRelocationCoordinate(
	value float64,
	field string,
) (float64, error) {
	if math.IsNaN(value) ||
		math.IsInf(value, 0) {
		return 0,
			&CharacterPositionRelocationValidationError{
				Field:  field,
				Reason: "must be finite",
			}
	}

	if math.Trunc(value) != value {
		return 0,
			&CharacterPositionRelocationValidationError{
				Field:  field,
				Reason: "must identify an integral tile coordinate",
			}
	}

	if value <
		float64(minCharacterPositionRelocationCoordinate) ||
		value >
			float64(maxCharacterPositionRelocationCoordinate) {
		return 0,
			&CharacterPositionRelocationValidationError{
				Field: field,
				Reason: fmt.Sprintf(
					"must be within signed 32-bit range [%d,%d]",
					minCharacterPositionRelocationCoordinate,
					maxCharacterPositionRelocationCoordinate,
				),
			}
	}

	return value, nil
}
