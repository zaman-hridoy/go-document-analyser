package retrieval

import "github.com/google/uuid"

type Result struct {
	ChunkID      uuid.UUID `json:"chunk_id"`
	DocumentID   uuid.UUID `json:"document_id"`
	DocumentName string    `json:"document_name"`
	Content      string    `json:"content"`
	ChunkIndex   int       `json:"chunk_index"`
	PageNumber   *int      `json:"page_number"`
	Similarity   float64   `json:"similarity"`
}
