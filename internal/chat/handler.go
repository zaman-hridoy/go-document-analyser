package chat

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/zaman-hridoy/go-document-analyser/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

type createChatRequest struct {
	Title       string      `json:"title"`
	DocumentIDs []uuid.UUID `json:"document_ids"`
}

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req createChatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	chat, err := h.service.Create(r.Context(), req.Title, req.DocumentIDs)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, chat)
}
