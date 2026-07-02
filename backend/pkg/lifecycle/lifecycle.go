package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type CleanupFunc func(ctx context.Context) error

type Manager struct {
	cleanups []CleanupFunc
}

func NewManager() *Manager {
	return &Manager{
		cleanups: make([]CleanupFunc, 0),
	}
}

func (m *Manager) Register(fn CleanupFunc) {
	m.cleanups = append(m.cleanups, fn)
}

func (m *Manager) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	slog.Info("Shutdown signal received", "signal", sig.String())

	// Timeout de 10 segundos para encerramento gracioso
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Executa os cleanups na ordem inversa de registro
	for i := len(m.cleanups) - 1; i >= 0; i-- {
		cleanup := m.cleanups[i]
		if err := cleanup(ctx); err != nil {
			slog.Error("Error executing cleanup function", "error", err)
		}
	}
	slog.Info("All services cleaned up. Exit successful.")
}
