package document

import (
	"errors"
	"fmt"
	"net/http"

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

func (h *Handler) Upload(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize+(1<<20))

	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			httpx.Error(w, http.StatusRequestEntityTooLarge, "file exceeds 25 mb limit")
			return
		}

		httpx.Error(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "file is required")
		return
	}

	defer file.Close()

	doc, err := h.service.Upload(r.Context(), file, header)

	if err != nil {
		fmt.Println(err)
		switch {
		case errors.Is(err, ErrFileRequired):
			httpx.Error(w, http.StatusBadRequest, "file is required")
			return
		case errors.Is(err, ErrFileTooLarge):
			httpx.Error(w, http.StatusBadRequest, "file exceeds 25 MB limit")
			return
		case errors.Is(err, ErrUnsupportedFileType):
			httpx.Error(w, http.StatusBadRequest, "only PDF, TEXT and Markdown files are supported")
			return
		default:
			httpx.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	httpx.WriteJSON(w, http.StatusCreated, doc)
}
