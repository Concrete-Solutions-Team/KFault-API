package server

import (
	"net/http"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/middleware"
)

func (s *Server) MountEndpoints(repo *auth.Repository, authHandler *auth.Handler) {
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("what???"))
	})
	authMiddleware := middleware.AuthMiddleware(repo)
	// s.router.Get("/ws", handler.HandleWS)

	s.router.Post("/auth/register", authHandler.Register)
	s.router.Post("/auth/login", authHandler.Login)
	s.router.With(authMiddleware).Get("/auth/me", authHandler.Profile)
	s.router.With(authMiddleware).Post("/auth/logout", authHandler.LogOut)

}
