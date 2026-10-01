package ingestion

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) ReplaceChunks(
	ctx context.Context,
	documentID uuid.UUID,
	chunks []Chunk,
) error {
	// if len(chunks) == 0 {
	// 	return nil
	// }

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin chunk transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM document_chunks
		WHERE document_id = $1
		`,
		documentID,
	)

	if err != nil {
		return fmt.Errorf("delete old document chunks: %w", err)
	}

	const query = `
		INSERT INTO document_chunks (
			document_id,
			chunk_index,
			content,
			page_number,
			token_count,
			embedding
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		)
	`

	for _, chunk := range chunks {
		_, err := tx.Exec(
			ctx,
			query,
			documentID,
			chunk.ChunkIndex,
			chunk.Content,
			chunk.PageNumber,
			chunk.TokenCount,
			pgvector.NewVector(chunk.Embedding),
		)

		if err != nil {
			return fmt.Errorf("insert document chunk %d: %w", chunk.ChunkIndex, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit chunk transaction: %w", err)
	}

	return nil
}
