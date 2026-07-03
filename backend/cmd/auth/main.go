package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "strings"
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
    Success   bool   `json:"success"`
    Token     string `json:"token,omitempty"`
    AccountID int    `json:"account_id,omitempty"`
    Error     string `json:"error,omitempty"`
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

func (s *AuthServer) authenticateAccount(ctx context.Context, username string, password string) (int, error) {
    username = strings.TrimSpace(username)
    password = strings.TrimSpace(password)

    if username == "" || password == "" {
        return 0, fmt.Errorf("username and password are required")
    }

    if s.pgPool == nil || s.pgPool.DB == nil {
        slog.Warn("PostgreSQL unavailable in Auth Server. Falling back to default account_id=1", "username", username)
        return 1, nil
    }

    var accountID int
    var passwordHash string

    err := s.pgPool.DB.QueryRowContext(ctx, `
        SELECT id, password_hash
        FROM accounts
        WHERE username = $1
    `, username).Scan(&accountID, &passwordHash)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return 0, fmt.Errorf("invalid credentials")
        }

        return 0, fmt.Errorf("failed to query account: %w", err)
    }

    // TODO FASE 3.3+: validar passwordHash com bcrypt.
    // Por enquanto, a senha precisa apenas estar preenchida e a conta precisa existir.
    _ = passwordHash

    return accountID, nil
}

func (s *AuthServer) startServer() {
    mux := http.NewServeMux()

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "UP", "service": "auth"}`))
    })

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

        req.Username = strings.TrimSpace(req.Username)
        req.Password = strings.TrimSpace(req.Password)

        slog.Info("Processing auth request", "username", req.Username)

        authCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        accountID, err := s.authenticateAccount(authCtx, req.Username, req.Password)
        if err != nil {
            slog.Warn("Auth request rejected", "username", req.Username, "error", err)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(AuthResponse{
                Success: false,
                Error:   "invalid_credentials",
            })
            return
        }

        token := fmt.Sprintf("session_%d", time.Now().UnixNano())

        if s.redisClient != nil {
            sessionValue := fmt.Sprintf("%d|%s", accountID, req.Username)
            err := s.redisClient.Client.Set(authCtx, token, sessionValue, 2*time.Hour).Err()
            if err != nil {
                slog.Error("Failed to store session in Redis", "error", err)
            }
        }

        messaging.GetInstance().Publish("auth.login", req.Username)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(AuthResponse{
            Success:   true,
            Token:     token,
            AccountID: accountID,
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
