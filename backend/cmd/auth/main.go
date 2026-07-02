package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/light-and-shadow/backend/config"
	"github.com/light-and-shadow/backend/pkg/db"
	"github.com/light-and-shadow/backend/pkg/lifecycle"
	"github.com/light-and-shadow/backend/pkg/logger"
	"github.com/light-and-shadow/backend/pkg/messaging"
)

type AuthServer struct {
	config      *config.Config
	httpServer  *http.Server
	pgPool      *db.PostgresPool
	redisClient *db.RedisClient
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.LogLevel)

	slog.Info("Starting Light and Shadow Auth Server...")

	pgPool, err := db.NewPostgresPool(cfg.PostgresDSN)
	if err != nil {
		slog.Warn("PostgreSQL connection failed in Auth Server", "error", err)
	}

	redisClient, err := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Warn("Redis connection failed in Auth Server", "error", err)
	}

	server := &AuthServer{
		config:      cfg,
		pgPool:      pgPool,
		redisClient: redisClient,
	}

	lifecycleMgr := lifecycle.NewManager()

	server.startServer()

	lifecycleMgr.Register(server.Shutdown)
	lifecycleMgr.Wait()
}

func (s *AuthServer) startServer() {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP", "service": "auth"}`))
	})

	// Internal Auth API endpoint
	mux.HandleFunc("/api/v1/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		slog.Info("Processing auth request", "username", req.Username)

		// Simulação de verificação segura e geração de token de sessão
		token := fmt.Sprintf("session_%d", time.Now().UnixNano())

		if s.redisClient != nil {
			// Grava a sessão ativa com TTL de 2 horas no cache (PATCH 3)
			err := s.redisClient.Client.Set(context.Background(), token, req.Username, 2*time.Hour).Err()
			if err != nil {
				slog.Error("Failed to store session in Redis", "error", err)
			}
		}

		// Publicação no Barramento Interno de Mensagens (PATCH 1)
		messaging.GetInstance().Publish("auth.login", req.Username)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Success: true,
			Token:   token,
		})
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.AuthPort),
		Handler: mux,
	}

	go func() {
		slog.Info("Auth Server listening", "port", s.config.AuthPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Auth HTTP Server error", "error", err)
		}
	}()
}

func (s *AuthServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down Auth Server gracefully...")

	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}

	if s.pgPool != nil {
		s.pgPool.Close(ctx)
	}

	if s.redisClient != nil {
		s.redisClient.Close(ctx)
	}

	slog.Info("Auth Server shutdown complete.")
	return nil
}
