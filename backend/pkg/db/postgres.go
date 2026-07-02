package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type PostgresPool struct {
	DB *sql.DB
}

func NewPostgresPool(dsn string) (*PostgresPool, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configuração do pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Teste de conexão (Ping)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	slog.Info("PostgreSQL connection pool initialized successfully")
	return &PostgresPool{DB: db}, nil
}

func (p *PostgresPool) Close(ctx context.Context) error {
	slog.Info("Closing PostgreSQL connection pool...")
	return p.DB.Close()
}
