package db

import (
	"context"
	"fmt"

	"github.com/AleeCao/LogiTrack/pkg/config"
	"github.com/jackc/pgx/v5"
)

func NewDBConnection(ctx context.Context, env *config.EnvConfig) (*pgx.Conn, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env.DBUser, env.DBPassword, env.DBHost, env.DBPort, env.DBName, env.DBSsl)

	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("error initializing db: %w", err)
	}

	return db, nil
}
