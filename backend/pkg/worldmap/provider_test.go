package worldmap

import (
	"sync"
	"testing"
)

func TestDebugProvider(t *testing.T) {
	provider := NewDebugProvider()

	if provider.Mode() != ModeDebug {
		t.Errorf("Expected mode to be '%s', got '%s'", ModeDebug, provider.Mode())
	}

	if provider.WorldID() != "alpha_debug_world" {
		t.Errorf("Expected world ID to be 'alpha_debug_world', got '%s'", provider.WorldID())
	}

	if provider.Version() != 1 {
		t.Errorf("Expected version to be 1, got %d", provider.Version())
	}
}

func TestDebugProvider_Concurrency(t *testing.T) {
	provider := NewDebugProvider()
	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = provider.Mode()
			_ = provider.WorldID()
			_ = provider.Version()
		}()
	}

	wg.Wait()
	// The test passes if it completes without a data race.
	// Run with -race flag to verify.
}
