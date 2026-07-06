package charactercreation

import (
	"context"
	"errors"
	"testing"

	"github.com/light-and-shadow/backend/pkg/persistence"
)

// fakeCreator is a mock implementation of the CharacterCreator interface for testing.
type fakeCreator struct {
	// Control what the mock returns
	retSummary   *persistence.CharacterSummary
	retErrorCode string
	retError     error

	// Capture what was passed to the mock
	calledWithCtx       context.Context
	calledWithAccountID int
	calledWithName      string
	calledWithRaceID    string
}

func (f *fakeCreator) CreateCharacterForAccount(ctx context.Context, accountID int, desiredName string, raceID string) (*persistence.CharacterSummary, string, error) {
	f.calledWithCtx = ctx
	f.calledWithAccountID = accountID
	f.calledWithName = desiredName
	f.calledWithRaceID = raceID
	return f.retSummary, f.retErrorCode, f.retError
}

// fakeRaceValidator is a mock implementation of the RaceValidator interface for testing.
type fakeRaceValidator struct {
	playableRaces map[string]bool
}

func (v *fakeRaceValidator) IsPlayableRace(raceID string) bool {
	if v.playableRaces == nil {
		return false
	}
	return v.playableRaces[raceID]
}

func TestCreateCharacter_Validation(t *testing.T) {
	creator := &fakeCreator{}
	validator := &fakeRaceValidator{
		playableRaces: map[string]bool{"human": true, "dwarf": true},
	}
	service := NewService(creator, validator)
	ctx := context.Background()

	testCases := []struct {
		name          string
		accountID     int
		req           CreateRequest
		expectedCode  string
		expectedError bool
	}{
		{"Invalid accountID", 0, CreateRequest{DesiredName: "Valid", RaceID: "human"}, "not_authenticated", true},
		{"Empty desiredName", 1, CreateRequest{DesiredName: "", RaceID: "human"}, "invalid_name", true},
		{"Whitespace desiredName", 1, CreateRequest{DesiredName: "   ", RaceID: "human"}, "invalid_name", true},
		{"Empty raceID", 1, CreateRequest{DesiredName: "Valid", RaceID: ""}, "invalid_race", true},
		{"Whitespace raceID", 1, CreateRequest{DesiredName: "Valid", RaceID: "  "}, "invalid_race", true},
		{"Unplayable race", 1, CreateRequest{DesiredName: "Valid", RaceID: "ogre"}, "invalid_race", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, code, err := service.CreateCharacter(ctx, tc.accountID, tc.req)
			if code != tc.expectedCode {
				t.Errorf("expected error code %q, got %q", tc.expectedCode, code)
			}
			if (err != nil) != tc.expectedError {
				t.Errorf("expected error presence to be %v, but was %v (err: %v)", tc.expectedError, (err != nil), err)
			}
		})
	}
}

func TestCreateCharacter_Dependencies(t *testing.T) {
	ctx := context.Background()
	req := CreateRequest{DesiredName: "Valid", RaceID: "human"}
	creator := &fakeCreator{}
	validator := &fakeRaceValidator{
		playableRaces: map[string]bool{"human": true},
	}

	t.Run("Nil creator", func(t *testing.T) {
		service := NewService(nil, validator)
		_, code, err := service.CreateCharacter(ctx, 1, req)
		if code != "internal_error" {
			t.Errorf("expected 'internal_error', got %q", code)
		}
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("Nil race validator", func(t *testing.T) {
		service := NewService(creator, nil)
		_, code, err := service.CreateCharacter(ctx, 1, req)
		if code != "internal_error" {
			t.Errorf("expected 'internal_error', got %q", code)
		}
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestCreateCharacter_Success(t *testing.T) {
	creator := &fakeCreator{
		retSummary: &persistence.CharacterSummary{
			ID:    123,
			Name:  "MyChar",
			Class: "novice",
			Level: 1,
		},
	}
	validator := &fakeRaceValidator{
		playableRaces: map[string]bool{"human": true},
	}
	service := NewService(creator, validator)
	ctx := context.Background()

	req := CreateRequest{
		DesiredName: "  MyChar  ",
		RaceID:      "  human  ",
	}

	summary, code, err := service.CreateCharacter(ctx, 42, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != "" {
		t.Fatalf("expected empty error code, got %q", code)
	}
	if summary == nil {
		t.Fatal("expected a character summary, got nil")
	}

	if summary.ID != 123 {
		t.Errorf("expected summary ID 123, got %d", summary.ID)
	}

	// Verify that normalized values were passed to the creator
	if creator.calledWithAccountID != 42 {
		t.Errorf("expected creator to be called with accountID 42, got %d", creator.calledWithAccountID)
	}
	if creator.calledWithName != "MyChar" {
		t.Errorf("expected creator to be called with normalized name 'MyChar', got %q", creator.calledWithName)
	}
	if creator.calledWithRaceID != "human" {
		t.Errorf("expected creator to be called with normalized raceID 'human', got %q", creator.calledWithRaceID)
	}
}

func TestCreateCharacter_PersistenceErrorHandling(t *testing.T) {
	validator := &fakeRaceValidator{
		playableRaces: map[string]bool{"human": true},
	}
	ctx := context.Background()
	req := CreateRequest{DesiredName: "Existing", RaceID: "human"}

	t.Run("Persistence returns specific error code", func(t *testing.T) {
		creator := &fakeCreator{
			retErrorCode: "name_taken",
			retError:     errors.New("db unique constraint violation"),
		}
		service := NewService(creator, validator)

		_, code, err := service.CreateCharacter(ctx, 1, req)

		if code != "name_taken" {
			t.Errorf("expected propagated error code 'name_taken', got %q", code)
		}
		if err == nil {
			t.Error("expected an error to be returned")
		}
	})

	t.Run("Persistence returns generic error", func(t *testing.T) {
		creator := &fakeCreator{
			retErrorCode: "", // No specific code provided
			retError:     errors.New("db connection lost"),
		}
		service := NewService(creator, validator)

		_, code, err := service.CreateCharacter(ctx, 1, req)

		if code != "internal_error" {
			t.Errorf("expected 'internal_error' for generic persistence failure, got %q", code)
		}
		if err == nil {
			t.Error("expected an error to be returned")
		}
	})
}
