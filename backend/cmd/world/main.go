package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/light-and-shadow/backend/config"
	"github.com/light-and-shadow/backend/pkg/db"
	"github.com/light-and-shadow/backend/pkg/lifecycle"
	"github.com/light-and-shadow/backend/pkg/logger"
	"github.com/light-and-shadow/backend/pkg/messaging"
	"github.com/light-and-shadow/backend/pkg/scheduler"
)

type WorldServer struct {
	config          *config.Config
	httpServer      *http.Server
	pgPool          *db.PostgresPool
	redisClient     *db.RedisClient
	tickScheduler   *scheduler.TickScheduler
	schedulerCancel context.CancelFunc
}

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.LogLevel)

	slog.Info("Starting Light and Shadow World Server...")

	pgPool, err := db.NewPostgresPool(cfg.PostgresDSN)
	if err != nil {
		slog.Warn("PostgreSQL connection failed in World Server", "error", err)
	}

	redisClient, err := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Warn("Redis connection failed in World Server", "error", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &WorldServer{
		config:          cfg,
		pgPool:          pgPool,
		redisClient:     redisClient,
		schedulerCancel: cancel,
	}

	lifecycleMgr := lifecycle.NewManager()

	server.startServer()
	server.startGameLoop(ctx)

	lifecycleMgr.Register(server.Shutdown)
	lifecycleMgr.Wait()
}

func (s *WorldServer) startServer() {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP", "service": "world"}`))
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.WorldPort),
		Handler: mux,
	}

	go func() {
		slog.Info("World Server listening", "port", s.config.WorldPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("World HTTP Server error", "error", err)
		}
	}()
}

func (s *WorldServer) startGameLoop(ctx context.Context) {
	// Cria o Scheduler de ticks de alta precisão configurado para 20Hz (PATCH 2)
	s.tickScheduler = scheduler.NewTickScheduler(20, func(dt time.Duration, tick uint64) {
		s.tick(dt, tick)
	})

	// Inicia a Game Loop de forma assíncrona
	go s.tickScheduler.Start(ctx)

	// Assina canais de mensagens de forma assíncrona desacoplada (PATCH 1)
	go func() {
		moveChan := messaging.GetInstance().Subscribe("player.move")
		for {
			select {
			case msg, ok := <-moveChan:
				if !ok {
					return
				}
				slog.Debug("World Server processed decoupled player move via Message Bus", "data", msg)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *WorldServer) tick(dt time.Duration, tick uint64) {
	// Simulação física de MMORPG baseada em delta time real (PATCH 2)
	// slog.Debug("Game Loop StepExecuted", "tick", tick, "dt", dt.Seconds())
}

func (s *WorldServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down World Server gracefully...")

	// Cancela contexto do Scheduler de Ticks
	if s.schedulerCancel != nil {
		s.schedulerCancel()
	}

	if s.tickScheduler != nil {
		s.tickScheduler.Stop()
	}

	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}

	if s.pgPool != nil {
		s.pgPool.Close(ctx)
	}

	if s.redisClient != nil {
		s.redisClient.Close(ctx)
	}

	slog.Info("World Server shutdown complete.")
	return nil
}
