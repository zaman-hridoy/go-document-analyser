package retrieval

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

type searchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	results, err := h.service.Search(r.Context(), req.Query, []uuid.UUID{}, req.Limit)

	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	httpx.WriteJSON(
		w,
		http.StatusOK,
		map[string]any{
			"results": results,
		},
	)
}
