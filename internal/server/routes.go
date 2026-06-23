package server

import (
	"net/http"
)

func (s *Server) MountEndpoints() {
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("what???"))
	})
	// s.router.Get("/ws", handler.HandleWS)

	// s.router.Post("/auth/register", handler.HandleRegister)
}
