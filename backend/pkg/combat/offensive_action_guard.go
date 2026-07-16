package combat

import "errors"

const defaultOffensiveActionBlockedReason = "offensive action blocked"

// OffensiveActionValidator defines a generic precondition for initiating an
// offensive action. The caller decides where in the combat flow it is invoked.
type OffensiveActionValidator func(attackerID string) error

// ErrOffensiveActionBlocked reports that an offensive action was rejected
// before combat state was mutated.
type ErrOffensiveActionBlocked struct {
	Reason       string
	ShouldNotify bool
}

// Error implements the error interface.
func (e *ErrOffensiveActionBlocked) Error() string {
	if e == nil || e.Reason == "" {
		return defaultOffensiveActionBlockedReason
	}

	return e.Reason
}

// IsOffensiveActionBlocked reports whether err contains an
// ErrOffensiveActionBlocked, including through wrapped errors.
func IsOffensiveActionBlocked(err error) bool {
	if err == nil {
		return false
	}

	var target *ErrOffensiveActionBlocked
	return errors.As(err, &target)
}