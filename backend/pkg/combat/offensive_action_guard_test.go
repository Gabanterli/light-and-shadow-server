package combat

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrOffensiveActionBlocked(t *testing.T) {
	t.Run("explicit reason", func(t *testing.T) {
		const reason = "test reason"

		err := &ErrOffensiveActionBlocked{
			Reason: reason,
		}

		if got := err.Error(); got != reason {
			t.Fatalf("Error() = %q, want %q", got, reason)
		}
	})

	t.Run("empty reason uses fallback", func(t *testing.T) {
		err := &ErrOffensiveActionBlocked{}

		if got := err.Error(); got != defaultOffensiveActionBlockedReason {
			t.Fatalf(
				"Error() = %q, want %q",
				got,
				defaultOffensiveActionBlockedReason,
			)
		}
	})

	t.Run("nil receiver uses fallback", func(t *testing.T) {
		var err *ErrOffensiveActionBlocked

		if got := err.Error(); got != defaultOffensiveActionBlockedReason {
			t.Fatalf(
				"Error() = %q, want %q",
				got,
				defaultOffensiveActionBlockedReason,
			)
		}
	})

	t.Run("ShouldNotify true is preserved", func(t *testing.T) {
		err := &ErrOffensiveActionBlocked{
			ShouldNotify: true,
		}

		if !err.ShouldNotify {
			t.Fatal("ShouldNotify = false, want true")
		}
	})

	t.Run("ShouldNotify false is preserved", func(t *testing.T) {
		err := &ErrOffensiveActionBlocked{
			ShouldNotify: false,
		}

		if err.ShouldNotify {
			t.Fatal("ShouldNotify = true, want false")
		}
	})
}

func TestIsOffensiveActionBlocked(t *testing.T) {
	t.Run("direct typed error", func(t *testing.T) {
		err := &ErrOffensiveActionBlocked{}

		if !IsOffensiveActionBlocked(err) {
			t.Fatal("IsOffensiveActionBlocked() = false, want true")
		}
	})

	t.Run("wrapped typed error", func(t *testing.T) {
		err := fmt.Errorf(
			"wrapped: %w",
			&ErrOffensiveActionBlocked{},
		)

		if !IsOffensiveActionBlocked(err) {
			t.Fatal("IsOffensiveActionBlocked() = false, want true")
		}
	})

	t.Run("errors As finds typed error", func(t *testing.T) {
		err := fmt.Errorf(
			"wrapped: %w",
			&ErrOffensiveActionBlocked{
				Reason:       "blocked",
				ShouldNotify: true,
			},
		)

		var blocked *ErrOffensiveActionBlocked

		if !errors.As(err, &blocked) {
			t.Fatal("errors.As() = false, want true")
		}

		if blocked.Reason != "blocked" {
			t.Fatalf(
				"Reason = %q, want %q",
				blocked.Reason,
				"blocked",
			)
		}

		if !blocked.ShouldNotify {
			t.Fatal("ShouldNotify = false, want true")
		}
	})

	t.Run("different error", func(t *testing.T) {
		err := errors.New("different error")

		if IsOffensiveActionBlocked(err) {
			t.Fatal("IsOffensiveActionBlocked() = true, want false")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if IsOffensiveActionBlocked(nil) {
			t.Fatal("IsOffensiveActionBlocked(nil) = true, want false")
		}
	})
}
