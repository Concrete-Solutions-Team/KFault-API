package storage

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	s *Storage
}

func NewHandler(s *Storage) *Handler {
	return &Handler{
		s: s,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	urls, _ := h.s.ListFileLinks(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}

func (h *Handler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	var data FileDataRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalig request", http.StatusBadRequest)
		return
	}

	url, err := h.s.GetUploadURL(r.Context(), data)
	if err != nil {
		http.Error(w, "Could not get upload URL", http.StatusInternalServerError)
		return
	}
	resp := UploadResponse{UploadUrl: url}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	var data FileDataRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalig request", http.StatusBadRequest)
		return
	}

	if err := h.s.DeleteFile(r.Context(), data.Key); err != nil {
		http.Error(w, "Could not delete file", http.StatusInternalServerError)
		return
	}

}
