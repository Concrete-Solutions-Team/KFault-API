package rooms

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/helpers"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body RoomData
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	ai, err := auth.GetAuthInfo(r)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	room := &RoomData{
		Name:    body.Name,
		OwnerID: ai.Claims.ID,
	}
	fmt.Println(ai.Claims.UserID)
	rd, err := h.service.CreateRoom(ctx, room, &ai.Claims.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
		return
	}

	helpers.SendJSON(w, http.StatusOK, rd)
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body RoomData
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	ai, err := auth.GetAuthInfo(r)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	room := &RoomData{
		Name:    body.Name,
		OwnerID: ai.Claims.ID,
	}
	fmt.Println(ai.Claims.UserID)
	rd, err := h.service.CreateRoom(ctx, room, &ai.Claims.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
		return
	}

	helpers.SendJSON(w, http.StatusOK, rd)
}

func (h *Handler) GetAllRooms(w http.ResponseWriter, r *http.Request)  {
	rooms, err := h.service.GetAllRooms(r.Context())
	if err != nil {
		helpers.SendJSON(w, http.StatusBadRequest, err)
		return
	}
	helpers.SendJSON(w, http.StatusOK, rooms)
}
