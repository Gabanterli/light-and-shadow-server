package combat

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var errSentinel = errors.New("offensive action blocked by test sentinel")

type fakeOffensiveActionValidator struct {
	mu         sync.Mutex
	shouldFail map[string]bool
	callCount  map[string]int
	onCall     func(attackerID string)
}

func newFakeOffensiveActionValidator() *fakeOffensiveActionValidator {
	return &fakeOffensiveActionValidator{
		shouldFail: make(map[string]bool),
		callCount:  make(map[string]int),
	}
}

func (f *fakeOffensiveActionValidator) Validate(attackerID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.callCount[attackerID]++

	if f.onCall != nil {
		f.onCall(attackerID)
	}

	if f.shouldFail[attackerID] {
		return errSentinel
	}
	return nil
}

func (f *fakeOffensiveActionValidator) SetShouldFail(attackerID string, fail bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shouldFail[attackerID] = fail
}

func (f *fakeOffensiveActionValidator) CallCount(attackerID string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount[attackerID]
}

func setupGuardTest(t *testing.T) (*CombatManager, *fakeOffensiveActionValidator) {
	t.Helper()

	// Load skills from the canonical config file to ensure tests use real data.
	// We navigate up from the current test file's directory.
	// This is brittle but avoids hardcoding absolute paths.
	// A better solution would be a testdata package or build tags.
	skillPath := filepath.Join("..", "..", "config", "alpha_skills.json")
	skills, err := LoadAlphaSkills(skillPath)
	if err != nil {
		t.Fatalf("Failed to load skills for test setup: %v", err)
	}

	cm := NewCombatManager(nil, skills)
	t.Cleanup(cm.Close)
	validator := newFakeOffensiveActionValidator()
	cm.SetOffensiveActionValidator(validator.Validate)

	// Register entities
	cm.RegisterEntity(&EntityStats{ID: "player1", Level: 10, Health: 100, MaxHealth: 100, Mana: 50, MaxMana: 50}, 0, 0)
	cm.RegisterEntity(&EntityStats{ID: "player2", Level: 10, Health: 100, MaxHealth: 100, Mana: 50, MaxMana: 50}, 1, 0)

	return cm, validator
}

func TestSkillCategory_RequiresOffensiveAuthorization(t *testing.T) {
	testCases := []struct {
		name     string
		category SkillCategory
		want     bool
	}{
		{"offensive", SkillCategoryOffensive, true},
		{"support", SkillCategorySupport, false},
		{"defensive", SkillCategoryDefensive, false},
		{"utility", SkillCategoryUtility, false},
		{"empty", "", true},
		{"unknown", "unknown_category", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.category.RequiresOffensiveAuthorization(); got != tc.want {
				t.Errorf("RequiresOffensiveAuthorization() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadAlphaSkills_CategoryValidation(t *testing.T) {
	validContent := `{
		"Skills": [
			{"ID": 1, "Name": "Offensive", "Category": "offensive", "CooldownMs": 1000, "Range": 1, "SkillScale": 1},
			{"ID": 2, "Name": "Support", "Category": "support", "CooldownMs": 1000, "Range": 1, "SkillScale": 1},
			{"ID": 3, "Name": "Defensive", "Category": "defensive", "CooldownMs": 1000, "Range": 1, "SkillScale": 1},
			{"ID": 4, "Name": "Utility", "Category": "utility", "CooldownMs": 1000, "Range": 1, "SkillScale": 1}
		]
	}`

	testCases := []struct {
		name        string
		content     string
		expectError bool
		errorText   string
	}{
		{"valid categories", validContent, false, ""},
		{"missing category", `{"Skills": [{"ID": 1, "Name": "Missing"}]}`, true, "missing category"},
		{"empty category", `{"Skills": [{"ID": 1, "Name": "Empty", "Category": ""}]}`, true, "missing category"},
		{"unknown category", `{"Skills": [{"ID": 1, "Name": "Unknown", "Category": "attack"}]}`, true, "unknown category"},
		{"wrong case", `{"Skills": [{"ID": 1, "Name": "Casing", "Category": "Offensive"}]}`, true, "unknown category"},
		{"with spaces", `{"Skills": [{"ID": 1, "Name": "Spaces", "Category": " offensive "}]}`, true, "unknown category"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "skills.json")
			if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write temp skills file: %v", err)
			}

			_, err := LoadAlphaSkills(path)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
				if !strings.Contains(err.Error(), tc.errorText) {
					t.Errorf("error %q does not contain expected text %q", err.Error(), tc.errorText)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
			}
		})
	}

	t.Run("canonical alpha skills are offensive", func(t *testing.T) {
		skillPath := filepath.Join("..", "..", "config", "alpha_skills.json")
		skills, err := LoadAlphaSkills(skillPath)
		if err != nil {
			t.Fatalf("Failed to load canonical alpha_skills.json: %v", err)
		}
		for _, id := range []uint32{1001, 1002, 1003} {
			skill, ok := skills[id]
			if !ok {
				t.Errorf("Skill %d not found in loaded map", id)
				continue
			}
			if skill.Category != SkillCategoryOffensive {
				t.Errorf("Skill %d has category %q, want %q", id, skill.Category, SkillCategoryOffensive)
			}
		}
	})
}

func TestCombatManager_SetOffensiveActionValidator(t *testing.T) {
	cm, _ := setupGuardTest(t)

	t.Run("set and replace", func(t *testing.T) {
		v1 := newFakeOffensiveActionValidator()
		cm.SetOffensiveActionValidator(v1.Validate)
		if cm.offensiveActionValidator == nil {
			t.Fatal("validator was not set")
		}

		v2 := newFakeOffensiveActionValidator()
		cm.SetOffensiveActionValidator(v2.Validate)
		if cm.offensiveActionValidator == nil {
			t.Fatal("validator was not replaced")
		}

		cm.SetOffensiveActionValidator(nil)
		if cm.offensiveActionValidator != nil {
			t.Fatal("validator was not cleared")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v := newFakeOffensiveActionValidator()
				cm.SetOffensiveActionValidator(v.Validate)
				cm.SetOffensiveActionValidator(nil)
			}()
		}
		wg.Wait()
	})
}

func TestCombatManager_Guard_Reentrancy(t *testing.T) {
	cm, validator := setupGuardTest(t)

	validator.onCall = func(attackerID string) {
		// Attempt to re-enter by calling another method on cm
		_, _ = cm.GetEntityStats(attackerID)
		// Attempt to replace the validator from within a validation call
		cm.SetOffensiveActionValidator(nil)
	}

	// Use a channel to signal completion and a timeout to detect deadlock
	done := make(chan struct{})
	go func() {
		_, _, _, _ = cm.ProcessAttackRequest("player1", "player2", "sword")
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("deadlock detected during re-entrant validator call")
	}

	if cm.offensiveActionValidator != nil {
		t.Error("validator should have been set to nil by the re-entrant call")
	}
}

func TestCombatManager_Guard_ProcessAttackRequest(t *testing.T) {
	cm, validator := setupGuardTest(t)
	attackerID, targetID := "player1", "player2"

	// Block the action
	validator.SetShouldFail(attackerID, true)

	// Set a cooldown to ensure guard runs first
	cm.nextAttackTime[attackerID] = time.Now().Add(1 * time.Hour)

	statsBefore, _ := cm.GetEntityStatsCopy(targetID)

	_, _, _, err := cm.ProcessAttackRequest(attackerID, targetID, "sword")

	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	if validator.CallCount(attackerID) != 1 {
		t.Fatalf("validator call count = %d, want 1", validator.CallCount(attackerID))
	}

	statsAfter, _ := cm.GetEntityStatsCopy(targetID)
	if statsAfter.Health != statsBefore.Health {
		t.Error("target health was mutated on a blocked action")
	}

	if _, exists := cm.lastAttacker[targetID]; exists {
		t.Error("lastAttacker was set on a blocked action")
	}
}

func TestCombatManager_Guard_ProcessCastSkillRequest(t *testing.T) {
	cm, validator := setupGuardTest(t)
	attackerID, targetID := "player1", "player2"

	// Create test skills
	offensiveSkill := Skill{ID: 9001, Name: "OffensiveTest", Category: SkillCategoryOffensive, Range: 5}
	supportSkill := Skill{ID: 9002, Name: "SupportTest", Category: SkillCategorySupport, Range: 5}
	defensiveSkill := Skill{ID: 9003, Name: "DefensiveTest", Category: SkillCategoryDefensive, Range: 5}
	utilitySkill := Skill{ID: 9004, Name: "UtilityTest", Category: SkillCategoryUtility, Range: 5}
	emptyCatSkill := Skill{ID: 9005, Name: "EmptyCatTest", Category: "", Range: 5}
	unknownCatSkill := Skill{ID: 9006, Name: "UnknownCatTest", Category: "unknown", Range: 5}

	cm.skills[offensiveSkill.ID] = offensiveSkill
	cm.skills[supportSkill.ID] = supportSkill
	cm.skills[defensiveSkill.ID] = defensiveSkill
	cm.skills[utilitySkill.ID] = utilitySkill
	cm.skills[emptyCatSkill.ID] = emptyCatSkill
	cm.skills[unknownCatSkill.ID] = unknownCatSkill

	validator.SetShouldFail(attackerID, true)

	t.Run("offensive skill is blocked", func(t *testing.T) {
		// Set cooldown and mana to ensure guard runs first
		cm.skillCooldowns[attackerID][offensiveSkill.ID] = time.Now().Add(1 * time.Hour)
		pStats, _ := cm.GetEntityStats(attackerID)
		pStats.Mana = 0

		_, err := cm.ProcessCastSkillRequest(attackerID, offensiveSkill.ID, targetID, 0, 0)
		if !errors.Is(err, errSentinel) {
			t.Fatalf("expected sentinel error, got %v", err)
		}
		if validator.CallCount(attackerID) != 1 {
			t.Fatalf("validator call count = %d, want 1", validator.CallCount(attackerID))
		}
	})

	validator.SetShouldFail(attackerID, false) // Reset for non-blocking tests

	nonOffensiveSkills := []Skill{supportSkill, defensiveSkill, utilitySkill}
	for _, skill := range nonOffensiveSkills {
		t.Run(fmt.Sprintf("%s skill is not blocked", skill.Category), func(t *testing.T) {
			callCountBefore := validator.CallCount(attackerID)
			_, err := cm.ProcessCastSkillRequest(attackerID, skill.ID, targetID, 0, 0)
			// We expect an error because ResolveSkill will fail for non-offensive skills without proper setup,
			// but it should NOT be the sentinel error.
			if errors.Is(err, errSentinel) {
				t.Fatal("non-offensive skill was blocked by guard")
			}
			callCountAfter := validator.CallCount(attackerID)
			if callCountAfter > callCountBefore {
				t.Fatal("validator was called for a non-offensive skill")
			}
		})
	}

	t.Run("fail-closed blocks empty and unknown categories", func(t *testing.T) {
		validator.SetShouldFail(attackerID, true)
		failClosedSkills := []Skill{emptyCatSkill, unknownCatSkill}
		for _, skill := range failClosedSkills {
			t.Run(string(skill.Category), func(t *testing.T) {
				callCountBefore := validator.CallCount(attackerID)
				_, err := cm.ProcessCastSkillRequest(attackerID, skill.ID, targetID, 0, 0)
				if !errors.Is(err, errSentinel) {
					t.Fatalf("expected sentinel error for category %q, got %v", skill.Category, err)
				}
				if validator.CallCount(attackerID) <= callCountBefore {
					t.Fatal("validator was not called for a fail-closed category")
				}
			})
		}
	})
}

func TestCombatManager_Guard_NoMutationOnBlock(t *testing.T) {
	cm, validator := setupGuardTest(t)
	attackerID, targetID := "player1", "player2"
	validator.SetShouldFail(attackerID, true)

	attackerStatsBefore, _ := cm.GetEntityStatsCopy(attackerID)
	targetStatsBefore, _ := cm.GetEntityStatsCopy(targetID)
	nextAttackTimeBefore := cm.nextAttackTime[attackerID]

	// Test basic attack
	_, _, isProj, err := cm.ProcessAttackRequest(attackerID, targetID, "sword")
	if err == nil {
		t.Fatal("attack was not blocked")
	}
	if isProj {
		t.Error("projectile was created for a blocked attack")
	}

	// Test skill cast
	offensiveSkillID := uint32(1001) // Fire Bolt
	_, err = cm.ProcessCastSkillRequest(attackerID, offensiveSkillID, targetID, 0, 0)
	if err == nil {
		t.Fatal("skill cast was not blocked")
	}

	// Verify no state was mutated
	attackerStatsAfter, _ := cm.GetEntityStatsCopy(attackerID)
	targetStatsAfter, _ := cm.GetEntityStatsCopy(targetID)
	nextAttackTimeAfter := cm.nextAttackTime[attackerID]

	if attackerStatsAfter.Health != attackerStatsBefore.Health ||
		attackerStatsAfter.Mana != attackerStatsBefore.Mana ||
		!attackerStatsAfter.LastCombatTime.IsZero() {
		t.Error("attacker stats were mutated")
	}

	if targetStatsAfter.Health != targetStatsBefore.Health {
		t.Error("target health was mutated")
	}

	if nextAttackTimeAfter != nextAttackTimeBefore {
		t.Error("basic attack cooldown was mutated")
	}

	if _, cdExists := cm.skillCooldowns[attackerID][offensiveSkillID]; cdExists {
		t.Error("skill cooldown was mutated")
	}
}

func TestCombatManager_Guard_ProjectileNotRevalidated(t *testing.T) {
	cm, validator := setupGuardTest(t)
	attackerID, targetID := "player1", "player2"

	validator.SetShouldFail(attackerID, true)

	resolved := make(chan struct{}, 1)
	cm.SetEventHandler(CombatEventHandler{
		OnDamage: func(atk, tgt string, dmg float64, isCrit, isHit bool, skillName string) {
			if atk == attackerID && tgt == targetID && skillName == "ProjectileGuardTest" {
				select {
				case resolved <- struct{}{}:
				default:
				}
			}
		},
	})

	p := &Projectile{
		ID:         "test_projectile_guard_no_revalidation",
		AttackerID: attackerID,
		TargetID:   targetID,
		TargetX:    1,
		TargetY:    0,
		SkillName:  "ProjectileGuardTest",
		Scaling:    1.0,
	}

	cm.LaunchProjectile(p, 10*time.Millisecond)

	select {
	case <-resolved:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for projectile resolution")
	}

	if got := validator.CallCount(attackerID); got != 0 {
		t.Fatalf("validator was called %d times during projectile resolution, want 0", got)
	}
}

func TestCombatManager_Guard_NilValidator(t *testing.T) {
	cm, _ := setupGuardTest(t)
	attackerID, targetID := "player1", "player2"

	// Disable the guard
	cm.SetOffensiveActionValidator(nil)

	// This action would otherwise be blocked, but should now succeed
	// (or fail for a different reason, like cooldown)
	_, _, _, err := cm.ProcessAttackRequest(attackerID, targetID, "sword")
	if errors.Is(err, errSentinel) {
		t.Fatal("action was blocked by a nil validator")
	}
}
