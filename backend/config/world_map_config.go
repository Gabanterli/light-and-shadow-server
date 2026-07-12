package config

import (
	"fmt"
	"strings"

	"github.com/light-and-shadow/backend/pkg/worldmap"
)

// ParseWorldMapMode validates and normalizes the configured world map mode.
func ParseWorldMapMode(rawMode string) (worldmap.Mode, error) {
	normalized := strings.ToLower(strings.TrimSpace(rawMode))

	switch normalized {
	case "":
		return worldmap.ModeDebug, nil
	case string(worldmap.ModeDebug):
		return worldmap.ModeDebug, nil
	case string(worldmap.ModeProduction):
		return worldmap.ModeProduction, nil
	default:
		return "", fmt.Errorf(
			"unknown world map mode %q: expected %q or %q",
			rawMode,
			worldmap.ModeDebug,
			worldmap.ModeProduction,
		)
	}
}
