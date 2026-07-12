package worldmap

// Mode represents the operational mode of the world map.
type Mode string

const (
	// ModeDebug uses the legacy, hardcoded procedural world.
	ModeDebug Mode = "debug"
	// ModeProduction uses the data-driven map files.
	ModeProduction Mode = "production"
)

// Provider is the authoritative source for static world map data.
type Provider interface {
	Mode() Mode
	WorldID() string
	Version() int
}
