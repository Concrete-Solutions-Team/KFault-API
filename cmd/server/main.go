package main

import (
	"log"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/config"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/db"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	pool := db.InitPostgres(cfg.DatabaseURL)

	// r.Post("/auth/register", handler.HandleRegister)
	// r.Post("/auth/login", handler.HandleLogin)
	// r.Post("/auth/me", handler.HandleMe)
	authRepository := auth.NewRepository(pool)
	authService := auth.NewService(authRepository)
	authHandler := auth.NewHandler(authService)

	s := server.NewServer(cfg.Port)
	s.MountEndpoints(authHandler)

	if err := s.Start(); err != nil {
		log.Println(err)
	}

}
