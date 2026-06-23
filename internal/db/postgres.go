package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres(dbURL string) *pgxpool.Pool {
	var pool *pgxpool.Pool
	ctx := context.Background()

	var err error
	pool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	log.Println("Connected to database successfully.")
	return pool
}
