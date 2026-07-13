package config

import (
	"os"
	"testing"
)

func unsetEnvironmentVariableForTest(
	t *testing.T,
	key string,
) {
	t.Helper()

	originalValue, existed := os.LookupEnv(key)

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("os.Unsetenv(%q) failed: %v", key, err)
	}

	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(key, originalValue); err != nil {
				t.Errorf(
					"os.Setenv(%q) cleanup failed: %v",
					key,
					err,
				)
			}
			return
		}

		if err := os.Unsetenv(key); err != nil {
			t.Errorf(
				"os.Unsetenv(%q) cleanup failed: %v",
				key,
				err,
			)
		}
	})
}

func TestLoadConfigWorldMapDefaults(t *testing.T) {
	unsetEnvironmentVariableForTest(
		t,
		"LS_WORLD_MAP_MODE",
	)
	unsetEnvironmentVariableForTest(
		t,
		"LS_WORLD_MAP_MANIFEST_PATH",
	)

	cfg := LoadConfig()

	if cfg.WorldMapMode != "debug" {
		t.Fatalf(
			"WorldMapMode = %q, want %q",
			cfg.WorldMapMode,
			"debug",
		)
	}

	const expectedPath = "config/worldmap/world_manifest.json"

	if cfg.WorldMapManifestPath != expectedPath {
		t.Fatalf(
			"WorldMapManifestPath = %q, want %q",
			cfg.WorldMapManifestPath,
			expectedPath,
		)
	}
}

func TestLoadConfigWorldMapOverrides(t *testing.T) {
	t.Setenv("LS_WORLD_MAP_MODE", "production")
	t.Setenv(
		"LS_WORLD_MAP_MANIFEST_PATH",
		"/srv/light-and-shadow/world_manifest.json",
	)

	cfg := LoadConfig()

	if cfg.WorldMapMode != "production" {
		t.Fatalf(
			"WorldMapMode = %q, want %q",
			cfg.WorldMapMode,
			"production",
		)
	}

	const expectedPath = "/srv/light-and-shadow/world_manifest.json"

	if cfg.WorldMapManifestPath != expectedPath {
		t.Fatalf(
			"WorldMapManifestPath = %q, want %q",
			cfg.WorldMapManifestPath,
			expectedPath,
		)
	}
}
