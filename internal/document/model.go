package document

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusUploaded   Status = "uploaded"
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusFailed     Status = "failed"
)

type Document struct {
	ID           uuid.UUID `json:"id"`
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"stored_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_sise"`
	StoragePath  string    `json:"storage_path"`
	Status       Status    `json:"status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
