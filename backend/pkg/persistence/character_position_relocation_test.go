package persistence

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

type fakeCharacterPositionRelocationDatabase struct {
	transaction characterPositionRelocationTransaction
	beginError  error
	beginCalls  int
	options     *sql.TxOptions
}

func (d *fakeCharacterPositionRelocationDatabase) BeginTx(
	_ context.Context,
	options *sql.TxOptions,
) (characterPositionRelocationTransaction, error) {
	d.beginCalls++

	if options != nil {
		copiedOptions := *options
		d.options = &copiedOptions
	}

	if d.beginError != nil {
		return nil, d.beginError
	}

	return d.transaction, nil
}

type fakeCharacterPositionRelocationTransaction struct {
	result sql.Result

	execError     error
	commitError   error
	rollbackError error

	query string
	args  []any

	execCalls     int
	commitCalls   int
	rollbackCalls int
}

func (t *fakeCharacterPositionRelocationTransaction) ExecContext(
	_ context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	t.execCalls++
	t.query = query
	t.args = append([]any(nil), args...)

	if t.execError != nil {
		return nil, t.execError
	}

	return t.result, nil
}

func (t *fakeCharacterPositionRelocationTransaction) Commit() error {
	t.commitCalls++
	return t.commitError
}

func (t *fakeCharacterPositionRelocationTransaction) Rollback() error {
	t.rollbackCalls++
	return t.rollbackError
}

type fakeCharacterPositionRelocationResult struct {
	rowsAffected      int64
	rowsAffectedError error
}

func (r *fakeCharacterPositionRelocationResult) LastInsertId() (
	int64,
	error,
) {
	return 0, errors.New(
		"LastInsertId is unsupported in this test",
	)
}

func (r *fakeCharacterPositionRelocationResult) RowsAffected() (
	int64,
	error,
) {
	if r.rowsAffectedError != nil {
		return 0, r.rowsAffectedError
	}

	return r.rowsAffected, nil
}

func validCharacterPositionRelocationRequest() CharacterPositionRelocationRequest {
	return CharacterPositionRelocationRequest{
		PlayerID:        "Gabriela",
		X:               100,
		Y:               101,
		Z:               2,
		ExpectedVersion: 7,
	}
}

func normalizedCharacterPositionRelocationRequest(
	t *testing.T,
	request CharacterPositionRelocationRequest,
) CharacterPositionRelocationRequest {
	t.Helper()

	normalized, err :=
		validateCharacterPositionRelocationRequest(
			request,
		)
	if err != nil {
		t.Fatalf(
			"validate relocation request: %v",
			err,
		)
	}

	return normalized
}

func relocationDatabaseWithRows(
	rowsAffected int64,
) (
	*fakeCharacterPositionRelocationDatabase,
	*fakeCharacterPositionRelocationTransaction,
) {
	transaction :=
		&fakeCharacterPositionRelocationTransaction{
			result: &fakeCharacterPositionRelocationResult{
				rowsAffected: rowsAffected,
			},
		}

	database :=
		&fakeCharacterPositionRelocationDatabase{
			transaction: transaction,
		}

	return database, transaction
}

func TestCharacterPositionRelocationValidation(
	t *testing.T,
) {
	testCases := []struct {
		name   string
		mutate func(*CharacterPositionRelocationRequest)
		field  string
	}{
		{
			name: "empty player",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.PlayerID = " "
			},
			field: "player_id",
		},
		{
			name: "nan x",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.X = math.NaN()
			},
			field: "x",
		},
		{
			name: "infinite y",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.Y = math.Inf(1)
			},
			field: "y",
		},
		{
			name: "fractional z",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.Z = 0.5
			},
			field: "z",
		},
		{
			name: "positive coordinate overflow",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.X = float64(1 << 31)
			},
			field: "x",
		},
		{
			name: "negative coordinate overflow",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.Y = float64((-1 << 31) - 1)
			},
			field: "y",
		},
		{
			name: "non-positive version",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.ExpectedVersion = 0
			},
			field: "expected_version",
		},
		{
			name: "version increment overflow",
			mutate: func(
				request *CharacterPositionRelocationRequest,
			) {
				request.ExpectedVersion = 1<<31 - 1
			},
			field: "expected_version",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				request :=
					validCharacterPositionRelocationRequest()

				testCase.mutate(&request)

				_, err :=
					validateCharacterPositionRelocationRequest(
						request,
					)
				if err == nil {
					t.Fatal(
						"expected relocation validation error",
					)
				}

				var validationError *CharacterPositionRelocationValidationError

				if !errors.As(err, &validationError) {
					t.Fatalf(
						"error type = %T, want CharacterPositionRelocationValidationError",
						err,
					)
				}

				if validationError.Field !=
					testCase.field {
					t.Fatalf(
						"validation field = %q, want %q",
						validationError.Field,
						testCase.field,
					)
				}
			},
		)
	}
}

func TestCharacterPositionRelocationRequiresDatabase(
	t *testing.T,
) {
	manager := NewPersistenceManager(nil)

	_, err := manager.RelocateCharacterPosition(
		context.Background(),
		validCharacterPositionRelocationRequest(),
	)
	if err == nil {
		t.Fatal(
			"expected persistence unavailable error",
		)
	}

	var unavailable *CharacterPositionRelocationPersistenceUnavailableError

	if !errors.As(err, &unavailable) {
		t.Fatalf(
			"error type = %T, want CharacterPositionRelocationPersistenceUnavailableError",
			err,
		)
	}
}

func TestCharacterPositionRelocationSuccess(
	t *testing.T,
) {
	database, transaction :=
		relocationDatabaseWithRows(1)

	request :=
		validCharacterPositionRelocationRequest()
	request.PlayerID = " Gabriela "

	normalizedRequest :=
		normalizedCharacterPositionRelocationRequest(
			t,
			request,
		)

	if normalizedRequest.PlayerID != "Gabriela" {
		t.Fatalf(
			"normalized player = %q, want Gabriela",
			normalizedRequest.PlayerID,
		)
	}

	result, err := executeCharacterPositionRelocation(
		context.Background(),
		database,
		normalizedRequest,
	)
	if err != nil {
		t.Fatalf(
			"relocate character position: %v",
			err,
		)
	}

	if result.NewVersion != 8 {
		t.Fatalf(
			"new version = %d, want 8",
			result.NewVersion,
		)
	}

	if database.beginCalls != 1 {
		t.Fatalf(
			"begin calls = %d, want 1",
			database.beginCalls,
		)
	}

	if database.options == nil ||
		database.options.Isolation !=
			sql.LevelRepeatableRead {
		t.Fatalf(
			"transaction options = %+v, want RepeatableRead",
			database.options,
		)
	}

	if transaction.execCalls != 1 ||
		transaction.commitCalls != 1 ||
		transaction.rollbackCalls != 0 {
		t.Fatalf(
			"exec=%d commit=%d rollback=%d, want 1/1/0",
			transaction.execCalls,
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}

	expectedArgs := []any{
		float64(100),
		float64(101),
		float64(2),
		"Gabriela",
		7,
	}

	if !reflect.DeepEqual(
		transaction.args,
		expectedArgs,
	) {
		t.Fatalf(
			"SQL args = %#v, want %#v",
			transaction.args,
			expectedArgs,
		)
	}

	normalizedQuery := strings.Join(
		strings.Fields(transaction.query),
		" ",
	)

	normalizedExpectedQuery := strings.Join(
		strings.Fields(
			characterPositionRelocationUpdateQuery,
		),
		" ",
	)

	if normalizedQuery != normalizedExpectedQuery {
		t.Fatalf(
			"query = %q, want %q",
			normalizedQuery,
			normalizedExpectedQuery,
		)
	}

	for _, forbidden := range []string{
		"inventories",
		"health",
		"mana",
		"experience",
		"gold",
		"class",
	} {
		if strings.Contains(
			strings.ToLower(normalizedQuery),
			forbidden,
		) {
			t.Fatalf(
				"relocation query contains forbidden field %q: %s",
				forbidden,
				normalizedQuery,
			)
		}
	}
}

func TestCharacterPositionRelocationConflict(
	t *testing.T,
) {
	database, transaction :=
		relocationDatabaseWithRows(0)

	request :=
		normalizedCharacterPositionRelocationRequest(
			t,
			validCharacterPositionRelocationRequest(),
		)

	_, err := executeCharacterPositionRelocation(
		context.Background(),
		database,
		request,
	)
	if err == nil {
		t.Fatal(
			"expected optimistic-lock conflict",
		)
	}

	var conflict *CharacterPositionRelocationConflictError

	if !errors.As(err, &conflict) {
		t.Fatalf(
			"error type = %T, want CharacterPositionRelocationConflictError",
			err,
		)
	}

	if conflict.PlayerID != "Gabriela" ||
		conflict.ExpectedVersion != 7 {
		t.Fatalf(
			"conflict = %+v",
			conflict,
		)
	}

	if transaction.commitCalls != 0 ||
		transaction.rollbackCalls != 1 {
		t.Fatalf(
			"commit=%d rollback=%d, want 0/1",
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
}

func TestCharacterPositionRelocationRejectsMultipleRows(
	t *testing.T,
) {
	database, transaction :=
		relocationDatabaseWithRows(2)

	request :=
		normalizedCharacterPositionRelocationRequest(
			t,
			validCharacterPositionRelocationRequest(),
		)

	_, err := executeCharacterPositionRelocation(
		context.Background(),
		database,
		request,
	)
	if err == nil {
		t.Fatal(
			"expected integrity error",
		)
	}

	var integrity *CharacterPositionRelocationIntegrityError

	if !errors.As(err, &integrity) {
		t.Fatalf(
			"error type = %T, want CharacterPositionRelocationIntegrityError",
			err,
		)
	}

	if integrity.RowsAffected != 2 {
		t.Fatalf(
			"rows affected = %d, want 2",
			integrity.RowsAffected,
		)
	}

	if transaction.commitCalls != 0 ||
		transaction.rollbackCalls != 1 {
		t.Fatalf(
			"commit=%d rollback=%d, want 0/1",
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
}

func TestCharacterPositionRelocationBeginFailure(
	t *testing.T,
) {
	beginFailure := errors.New(
		"begin failed",
	)

	database :=
		&fakeCharacterPositionRelocationDatabase{
			beginError: beginFailure,
		}

	request :=
		normalizedCharacterPositionRelocationRequest(
			t,
			validCharacterPositionRelocationRequest(),
		)

	_, err := executeCharacterPositionRelocation(
		context.Background(),
		database,
		request,
	)
	if !errors.Is(err, beginFailure) {
		t.Fatalf(
			"error = %v, want wrapped begin failure",
			err,
		)
	}

	if database.beginCalls != 1 {
		t.Fatalf(
			"begin calls = %d, want 1",
			database.beginCalls,
		)
	}
}

func TestCharacterPositionRelocationExecuteFailure(
	t *testing.T,
) {
	executeFailure := errors.New(
		"execute failed",
	)

	transaction :=
		&fakeCharacterPositionRelocationTransaction{
			execError: executeFailure,
		}

	database :=
		&fakeCharacterPositionRelocationDatabase{
			transaction: transaction,
		}

	request :=
		normalizedCharacterPositionRelocationRequest(
			t,
			validCharacterPositionRelocationRequest(),
		)

	_, err := executeCharacterPositionRelocation(
		context.Background(),
		database,
		request,
	)
	if !errors.Is(err, executeFailure) {
		t.Fatalf(
			"error = %v, want wrapped execute failure",
			err,
		)
	}

	if transaction.commitCalls != 0 ||
		transaction.rollbackCalls != 1 {
		t.Fatalf(
			"commit=%d rollback=%d, want 0/1",
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
}

func TestCharacterPositionRelocationRowsAffectedFailure(
	t *testing.T,
) {
	rowsFailure := errors.New(
		"rows affected failed",
	)

	transaction :=
		&fakeCharacterPositionRelocationTransaction{
			result: &fakeCharacterPositionRelocationResult{
				rowsAffectedError: rowsFailure,
			},
		}

	database :=
		&fakeCharacterPositionRelocationDatabase{
			transaction: transaction,
		}

	request :=
		normalizedCharacterPositionRelocationRequest(
			t,
			validCharacterPositionRelocationRequest(),
		)

	_, err := executeCharacterPositionRelocation(
		context.Background(),
		database,
		request,
	)
	if !errors.Is(err, rowsFailure) {
		t.Fatalf(
			"error = %v, want wrapped rows failure",
			err,
		)
	}

	if transaction.commitCalls != 0 ||
		transaction.rollbackCalls != 1 {
		t.Fatalf(
			"commit=%d rollback=%d, want 0/1",
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
}

func TestCharacterPositionRelocationCommitFailure(
	t *testing.T,
) {
	commitFailure := errors.New(
		"commit failed",
	)

	transaction :=
		&fakeCharacterPositionRelocationTransaction{
			result: &fakeCharacterPositionRelocationResult{
				rowsAffected: 1,
			},
			commitError: commitFailure,
		}

	database :=
		&fakeCharacterPositionRelocationDatabase{
			transaction: transaction,
		}

	request :=
		normalizedCharacterPositionRelocationRequest(
			t,
			validCharacterPositionRelocationRequest(),
		)

	_, err := executeCharacterPositionRelocation(
		context.Background(),
		database,
		request,
	)
	if !errors.Is(err, commitFailure) {
		t.Fatalf(
			"error = %v, want wrapped commit failure",
			err,
		)
	}

	if transaction.commitCalls != 1 ||
		transaction.rollbackCalls != 1 {
		t.Fatalf(
			"commit=%d rollback=%d, want 1/1",
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
}

func TestCharacterPositionRelocationUsesCallerCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	database, _ :=
		relocationDatabaseWithRows(1)

	request :=
		normalizedCharacterPositionRelocationRequest(
			t,
			validCharacterPositionRelocationRequest(),
		)

	_, err := executeCharacterPositionRelocation(
		ctx,
		database,
		request,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"error = %v, want context.Canceled",
			err,
		)
	}

	if database.beginCalls != 0 {
		t.Fatalf(
			"begin calls = %d, want 0",
			database.beginCalls,
		)
	}
}
