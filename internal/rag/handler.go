package rag

import (
	"encoding/json"
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

type queryRequest struct {
	Question string `json:"question"`
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	chatID, err := uuid.Parse(r.PathValue("id"))

	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	var req queryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Query(r.Context(), chatID, req.Question)

	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) StreamQuery(w http.ResponseWriter, r *http.Request) {
	chatID, err := uuid.Parse(r.PathValue("id"))

	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	started := false

	err = h.service.StreamQuery(
		r.Context(),
		chatID,
		req.Question,
		func(metadata StreamMetadata) error {
			if !started {
				w.Header().Set("Content-Type", "text/event-stream")

				w.Header().Set("Cache-Control", "no-cache")

				w.Header().Set("Connection", "keep-alive")

				w.Header().Set("X-Accel-Buffering", "no")
				w.WriteHeader(http.StatusOK)

				started = true
			}

			return httpx.WriteSSE(w, "sources", metadata)
		},

		func(content string) error {
			return httpx.WriteSSE(
				w,
				"delta",
				map[string]string{
					"content": content,
				},
			)
		},
	)

	if err != nil {
		log.Println(err)
		if !started {
			httpx.Error(
				w,
				http.StatusInternalServerError,
				"failed to answer question",
			)
			return
		}

		_ = httpx.WriteSSE(
			w,
			"error",
			map[string]string{
				"message": "response generation failed",
			},
		)

		return
	}

	_ = httpx.WriteSSE(
		w,
		"done",
		map[string]bool{
			"done": true,
		},
	)

}
