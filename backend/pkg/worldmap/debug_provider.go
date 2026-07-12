package worldmap

// DebugProvider is an implementation of Provider that returns
// deterministic, hardcoded values for the Alpha/debug world.
type DebugProvider struct{}

// NewDebugProvider creates a new provider for the debug world.
func NewDebugProvider() *DebugProvider {
	return &DebugProvider{}
}

// Mode returns the map mode, which is always ModeDebug.
func (p *DebugProvider) Mode() Mode {
	return ModeDebug
}

// WorldID returns the hardcoded identifier for the debug world.
func (p *DebugProvider) WorldID() string {
	return "alpha_debug_world"
}

// Version returns the hardcoded version for the debug world.
func (p *DebugProvider) Version() int {
	return 1
}
