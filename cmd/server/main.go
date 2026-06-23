package main

import (
	"log"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/config"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	// pool := db.InitPostgres(cfg.DatabaseURL)

	r.Post("/auth/register", handler.HandleRegister)
	r.Post("/auth/login", handler.HandleLogin)
	r.Post("/auth/me", handler.HandleMe)
	s := server.NewServer(cfg.Port)
	s.MountEndpoints()

	if err := s.Start(); err != nil {
		log.Println(err)
	}

}
