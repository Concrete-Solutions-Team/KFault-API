package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/config"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/db"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/storage"
)

func CleanupExpiredTokens(ctx context.Context, db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := `SELECT cleanup_expired_tokens()`

	_, err := db.Exec(ctx, query)
	return err
}

func StartCleanup(ctx context.Context, db *pgxpool.Pool) {
	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := CleanupExpiredTokens(context.Background(), db); err != nil {
					log.Fatalf("Failed to clean up tokens: %v", err)
				}
				log.Println("Expired tokens cleaned up successfully!")
			case <-ctx.Done():
				fmt.Println("Stopping periodic task...")
				return
			}
		}
	}()
}

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithCancel(context.Background())

	str := storage.NewStorage(
		cfg.S3AccountID,
		cfg.S3AccessKeyID,
		cfg.S3AccessKeySecret,
		cfg.S3BucketName,
	)

	// str.SaveObjects()
	strHandler := storage.NewHandler(str)

	pool := db.InitPostgres(cfg.DatabaseURL)
	StartCleanup(ctx, pool)

	authRepository := auth.NewRepository(pool)
	authService := auth.NewService(authRepository)
	authHandler := auth.NewHandler(authService, cfg.FrontendURL)

	s := server.NewServer(cfg.Port, cfg.FrontendURL)
	s.MountEndpoints(authRepository, authHandler, strHandler)

	if err := s.Start(); err != nil {
		log.Println(err)
	}
	cancel()
	time.Sleep(1 * time.Second)
}
