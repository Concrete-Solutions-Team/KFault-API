package server

import (
	"net/http"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/middleware"
)

func (s *Server) MountEndpoints(authHandler *auth.Handler) {
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("what???"))
	})
	// s.router.Get("/ws", handler.HandleWS)

	s.router.Post("/auth/register", authHandler.Register)
	s.router.Post("/auth/login", authHandler.Login)
	s.router.With(middleware.AuthMiddleware).Post("/auth/me", authHandler.Profile)

}
