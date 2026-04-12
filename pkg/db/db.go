package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/AleeCao/LogiTrack/pkg/config"
	_ "github.com/lib/pq"
)

func NewDBConnection(env *config.EnvConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		env.DBUser, env.DBPassword, env.DBHost, env.DBPort, env.DBName, env.DBSsl)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error initializing db: %w", err)
	}

	db.SetMaxOpenConns(env.MaxConns)
	db.SetMaxIdleConns(env.MaxIdles)
	db.SetConnMaxLifetime(time.Duration(env.Lifetime) * time.Minute)

	return db, nil
}
