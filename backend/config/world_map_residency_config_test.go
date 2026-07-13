package config

import (
	"os"
	"testing"
)

func unsetEnvironmentForTest(
	t *testing.T,
	name string,
) {
	t.Helper()

	originalValue, originallyPresent := os.LookupEnv(name)

	if err := os.Unsetenv(name); err != nil {
		t.Fatalf(
			"unset environment variable %q: %v",
			name,
			err,
		)
	}

	t.Cleanup(func() {
		if originallyPresent {
			if err := os.Setenv(
				name,
				originalValue,
			); err != nil {
				t.Errorf(
					"restore environment variable %q: %v",
					name,
					err,
				)
			}

			return
		}

		if err := os.Unsetenv(name); err != nil {
			t.Errorf(
				"remove environment variable %q during cleanup: %v",
				name,
				err,
			)
		}
	})
}

func TestLoadConfigWorldMapInitialResidencyDefault(
	t *testing.T,
) {
	unsetEnvironmentForTest(
		t,
		"LS_WORLD_MAP_INITIAL_RESIDENCY_PATH",
	)

	cfg := LoadConfig()

	const expected = "config/worldmap/initial_residency.json"

	if cfg.WorldMapInitialResidencyPath != expected {
		t.Fatalf(
			"WorldMapInitialResidencyPath = %q, want %q",
			cfg.WorldMapInitialResidencyPath,
			expected,
		)
	}
}

func TestLoadConfigWorldMapInitialResidencyExplicitEmpty(
	t *testing.T,
) {
	t.Setenv(
		"LS_WORLD_MAP_INITIAL_RESIDENCY_PATH",
		"",
	)

	cfg := LoadConfig()

	if cfg.WorldMapInitialResidencyPath != "" {
		t.Fatalf(
			"WorldMapInitialResidencyPath = %q, want explicit empty value",
			cfg.WorldMapInitialResidencyPath,
		)
	}
}

func TestLoadConfigWorldMapInitialResidencyOverride(
	t *testing.T,
) {
	const expected = "/srv/light-and-shadow/initial_residency.json"

	t.Setenv(
		"LS_WORLD_MAP_INITIAL_RESIDENCY_PATH",
		expected,
	)

	cfg := LoadConfig()

	if cfg.WorldMapInitialResidencyPath != expected {
		t.Fatalf(
			"WorldMapInitialResidencyPath = %q, want %q",
			cfg.WorldMapInitialResidencyPath,
			expected,
		)
	}
}
