package config

import (
	"testing"

	"github.com/light-and-shadow/backend/pkg/worldmap"
)

func TestParseWorldMapMode(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    worldmap.Mode
		expectError bool
	}{
		{
			name:     "empty defaults to debug",
			input:    "",
			expected: worldmap.ModeDebug,
		},
		{
			name:     "whitespace defaults to debug",
			input:    "   ",
			expected: worldmap.ModeDebug,
		},
		{
			name:     "explicit debug",
			input:    "debug",
			expected: worldmap.ModeDebug,
		},
		{
			name:     "debug is case insensitive",
			input:    " DEBUG ",
			expected: worldmap.ModeDebug,
		},
		{
			name:     "explicit production",
			input:    "production",
			expected: worldmap.ModeProduction,
		},
		{
			name:     "production is case insensitive",
			input:    " PRODUCTION ",
			expected: worldmap.ModeProduction,
		},
		{
			name:        "unknown mode fails",
			input:       "invalid-mode",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mode, err := ParseWorldMapMode(testCase.input)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if mode != testCase.expected {
				t.Fatalf(
					"unexpected mode: got %q, expected %q",
					mode,
					testCase.expected,
				)
			}
		})
	}
}
