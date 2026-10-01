package ingestion

import (
	"log"
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

func (h *Handler) ProcessDocument(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")

	documentID, err := uuid.Parse(idValue)
	if err != nil {
		httpx.Error(
			w,
			http.StatusBadRequest,
			"invalid document id",
		)
		return
	}

	err = h.service.Process(r.Context(), documentID)
	if err != nil {
		log.Println(err)
		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to process document",
		)
		return
	}

	httpx.WriteJSON(
		w,
		http.StatusOK,
		map[string]string{
			"status": "Processed",
		},
	)
}
