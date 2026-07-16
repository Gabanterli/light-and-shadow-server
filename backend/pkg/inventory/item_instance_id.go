package inventory

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

const (
	ItemInstanceIDPrefix                   = "item:"
	ItemInstanceIDEntropyBytes             = 16
	DefaultItemInstanceIDCollisionAttempts = 8
	MaxItemInstanceIDCollisionAttempts     = 1024
)

var (
	ErrItemInstanceIDInvalid          = errors.New("invalid item instance ID")
	ErrItemInstanceIDEntropy          = errors.New("item instance ID entropy failure")
	ErrItemInstanceIDAuthorityInvalid = errors.New("invalid item instance ID authority")
	ErrItemInstanceIDCollision        = errors.New("item instance ID collision limit reached")
	ErrItemInstanceIDAlreadyReserved  = errors.New("item instance ID is already reserved")
	ErrItemInstanceIDNotReserved      = errors.New("item instance ID is not reserved")
)

// ItemInstanceIDAuthority is the server-side reservation boundary for item
// instance identities. Entropy consumption and reservation publication are
// serialized so an ID is never returned before it is reserved.
type ItemInstanceIDAuthority struct {
	mu                   sync.RWMutex
	entropy              io.Reader
	maxCollisionAttempts int
	reserved             map[ItemInstanceID]struct{}
}

// NewItemInstanceIDAuthority creates the production authority backed by
// crypto/rand.Reader.
func NewItemInstanceIDAuthority() *ItemInstanceIDAuthority {
	authority, err := NewItemInstanceIDAuthorityWithEntropy(
		rand.Reader,
		DefaultItemInstanceIDCollisionAttempts,
	)
	if err != nil {
		panic("inventory: invalid built-in item instance ID authority configuration: " + err.Error())
	}
	return authority
}

// NewItemInstanceIDAuthorityWithEntropy creates a testable authority using the
// provided entropy source and bounded collision retries.
func NewItemInstanceIDAuthorityWithEntropy(
	entropy io.Reader,
	maxCollisionAttempts int,
) (*ItemInstanceIDAuthority, error) {
	if entropy == nil {
		return nil, fmt.Errorf("%w: entropy reader is required", ErrItemInstanceIDAuthorityInvalid)
	}
	if maxCollisionAttempts < 1 || maxCollisionAttempts > MaxItemInstanceIDCollisionAttempts {
		return nil, fmt.Errorf(
			"%w: collision attempts must be between 1 and %d",
			ErrItemInstanceIDAuthorityInvalid,
			MaxItemInstanceIDCollisionAttempts,
		)
	}
	return &ItemInstanceIDAuthority{
		entropy:              entropy,
		maxCollisionAttempts: maxCollisionAttempts,
		reserved:             make(map[ItemInstanceID]struct{}),
	}, nil
}

// GenerateItemInstanceID reads 128 bits from entropy and formats the opaque ID.
// It does not reserve the result; production callers should use ReserveNew.
func GenerateItemInstanceID(entropy io.Reader) (ItemInstanceID, error) {
	if entropy == nil {
		return "", fmt.Errorf("%w: entropy reader is required", ErrItemInstanceIDEntropy)
	}

	var bytes [ItemInstanceIDEntropyBytes]byte
	if _, err := io.ReadFull(entropy, bytes[:]); err != nil {
		return "", fmt.Errorf("%w: read %d bytes: %v", ErrItemInstanceIDEntropy, len(bytes), err)
	}

	encoded := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(encoded, bytes[:])
	id := ItemInstanceID(ItemInstanceIDPrefix + string(encoded))
	if err := ValidateItemInstanceID(id); err != nil {
		return "", fmt.Errorf("%w: generated identity failed validation: %v", ErrItemInstanceIDEntropy, err)
	}
	return id, nil
}

// ParseItemInstanceID converts a wire or persistence string into a strictly
// validated canonical identity.
func ParseItemInstanceID(value string) (ItemInstanceID, error) {
	id := ItemInstanceID(value)
	if err := ValidateItemInstanceID(id); err != nil {
		return "", err
	}
	return id, nil
}

// ValidateItemInstanceID enforces the canonical lowercase item:<32 hex> form.
func ValidateItemInstanceID(id ItemInstanceID) error {
	value := string(id)
	expectedLength := len(ItemInstanceIDPrefix) + (ItemInstanceIDEntropyBytes * 2)
	if len(value) != expectedLength {
		return fmt.Errorf(
			"%w: %q must contain %d characters",
			ErrItemInstanceIDInvalid,
			value,
			expectedLength,
		)
	}
	if !strings.HasPrefix(value, ItemInstanceIDPrefix) {
		return fmt.Errorf("%w: %q must start with %q", ErrItemInstanceIDInvalid, value, ItemInstanceIDPrefix)
	}

	hexPart := value[len(ItemInstanceIDPrefix):]
	if strings.ToLower(hexPart) != hexPart {
		return fmt.Errorf("%w: hexadecimal payload must be lowercase", ErrItemInstanceIDInvalid)
	}
	decoded, err := hex.DecodeString(hexPart)
	if err != nil || len(decoded) != ItemInstanceIDEntropyBytes {
		return fmt.Errorf("%w: hexadecimal payload must encode %d bytes", ErrItemInstanceIDInvalid, ItemInstanceIDEntropyBytes)
	}
	if err := validateCanonicalInventoryID("item instance ID", value); err != nil {
		return fmt.Errorf("%w: %v", ErrItemInstanceIDInvalid, err)
	}
	return nil
}

// ReserveNew generates and reserves a new ID before returning it. Collisions are
// retried up to the configured bound.
func (authority *ItemInstanceIDAuthority) ReserveNew() (ItemInstanceID, error) {
	if err := authority.validate(); err != nil {
		return "", err
	}

	authority.mu.Lock()
	defer authority.mu.Unlock()

	for attempt := 1; attempt <= authority.maxCollisionAttempts; attempt++ {
		id, err := GenerateItemInstanceID(authority.entropy)
		if err != nil {
			return "", err
		}
		if _, collision := authority.reserved[id]; collision {
			continue
		}
		authority.reserved[id] = struct{}{}
		return id, nil
	}

	return "", fmt.Errorf(
		"%w after %d attempts",
		ErrItemInstanceIDCollision,
		authority.maxCollisionAttempts,
	)
}

// ReserveExisting claims a validated identity loaded from an authoritative
// external source before that identity is published into runtime state.
func (authority *ItemInstanceIDAuthority) ReserveExisting(id ItemInstanceID) error {
	if err := authority.validate(); err != nil {
		return err
	}
	if err := ValidateItemInstanceID(id); err != nil {
		return err
	}

	authority.mu.Lock()
	defer authority.mu.Unlock()
	if _, exists := authority.reserved[id]; exists {
		return fmt.Errorf("%w: %q", ErrItemInstanceIDAlreadyReserved, id)
	}
	authority.reserved[id] = struct{}{}
	return nil
}

// Release removes exactly one reservation. Releasing an unknown or already
// released identity is rejected.
func (authority *ItemInstanceIDAuthority) Release(id ItemInstanceID) error {
	if err := authority.validate(); err != nil {
		return err
	}
	if err := ValidateItemInstanceID(id); err != nil {
		return err
	}

	authority.mu.Lock()
	defer authority.mu.Unlock()
	if _, exists := authority.reserved[id]; !exists {
		return fmt.Errorf("%w: %q", ErrItemInstanceIDNotReserved, id)
	}
	delete(authority.reserved, id)
	return nil
}

func (authority *ItemInstanceIDAuthority) IsReserved(id ItemInstanceID) (bool, error) {
	if err := authority.validate(); err != nil {
		return false, err
	}
	if err := ValidateItemInstanceID(id); err != nil {
		return false, err
	}

	authority.mu.RLock()
	_, exists := authority.reserved[id]
	authority.mu.RUnlock()
	return exists, nil
}

func (authority *ItemInstanceIDAuthority) Count() int {
	if authority == nil {
		return 0
	}
	authority.mu.RLock()
	count := len(authority.reserved)
	authority.mu.RUnlock()
	return count
}

// ReservedIDs returns a deterministic defensive snapshot.
func (authority *ItemInstanceIDAuthority) ReservedIDs() []ItemInstanceID {
	if authority == nil {
		return nil
	}
	authority.mu.RLock()
	ids := make([]ItemInstanceID, 0, len(authority.reserved))
	for id := range authority.reserved {
		ids = append(ids, id)
	}
	authority.mu.RUnlock()
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (authority *ItemInstanceIDAuthority) validate() error {
	if authority == nil {
		return fmt.Errorf("%w: authority is required", ErrItemInstanceIDAuthorityInvalid)
	}
	if authority.entropy == nil {
		return fmt.Errorf("%w: entropy reader is required", ErrItemInstanceIDAuthorityInvalid)
	}
	if authority.maxCollisionAttempts < 1 || authority.maxCollisionAttempts > MaxItemInstanceIDCollisionAttempts {
		return fmt.Errorf("%w: collision attempt configuration is invalid", ErrItemInstanceIDAuthorityInvalid)
	}
	if authority.reserved == nil {
		return fmt.Errorf("%w: reservation registry is not initialized", ErrItemInstanceIDAuthorityInvalid)
	}
	return nil
}
