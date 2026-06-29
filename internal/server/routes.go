package server

import (
	"net/http"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/messages"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/middleware"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/rooms"
)

func (s *Server) MountEndpoints(repo *auth.Repository, authHandler *auth.Handler, hub *messages.Hub, wsRepo *messages.Repository, roomsHandler *rooms.Handler) {
	"github.com/Concrete-Solutions-Team/KFault-API/internal/storage"
)

func (s *Server) MountEndpoints(repo *auth.Repository, authHandler *auth.Handler, strHandler *storage.Handler) {
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("what???"))
	})
	authMiddleware := middleware.AuthMiddleware(repo)
	// s.router.Get("/ws", handler.HandleWS)

	s.router.With(authMiddleware).HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		messages.ServeWS(hub, wsRepo, w, r)
	})
	s.router.HandleFunc("/w", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui/ws.html")
	})
	s.router.Get("/images/list", strHandler.List)
	s.router.Post("/images/upload/get-url", strHandler.GetUploadURL)
	s.router.Delete("/images/delete", strHandler.Delete)

	s.router.Post("/auth/register", authHandler.Register)
	s.router.Post("/auth/login", authHandler.Login)

	s.router.With(authMiddleware).Get("/auth/me", authHandler.Profile)
	s.router.With(authMiddleware).Post("/auth/logout", authHandler.LogOut)

	s.router.With(authMiddleware).Get("/rooms", roomsHandler.GetAllRooms)
	s.router.With(authMiddleware).Post("/rooms/create", roomsHandler.CreateRoom)

}
