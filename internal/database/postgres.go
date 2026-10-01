package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
)

func NewPostgresPool(
	ctx context.Context,
	databaseURL string,
) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)

	/*
		gets parsed into configuration containing things like:
			host
			port
			username
			password
			database
			SSL settings
			pool configuration
	*/
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if err := pgxvector.RegisterTypes(ctx, conn); err != nil {
			return fmt.Errorf("register pgvector types: %w", err)
		}
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
	}

	return pool, nil
}
