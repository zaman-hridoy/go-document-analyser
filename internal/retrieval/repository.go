package retrieval

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

func (r *Repository) Search(
	ctx context.Context,
	queryEmbedding []float32,
	documentIDs []uuid.UUID,
	limit int,
) ([]Result, error) {
	const query = `
		SELECT
			dc.id,
			dc.document_id,
			d.original_name,
			dc.content,
			dc.chunk_index,
			dc.page_number,

			1 - (
				dc.embedding <=> $1
			) AS similarity

		FROM document_chunks dc
		
		INNER JOIN documents d
			ON d.id = dc.document_id
		WHERE 
			d.status = 'ready'
			AND dc.embedding IS NOT NULL
			AND dc.document_id = ANY($2)
		ORDER BY
			dc.embedding <=> $1
		LIMIT $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		pgvector.NewVector(queryEmbedding),
		documentIDs,
		limit,
	)

	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	defer rows.Close()

	var results []Result

	for rows.Next() {
		var result Result

		if err := rows.Scan(
			&result.ChunkID,
			&result.DocumentID,
			&result.DocumentName,
			&result.Content,
			&result.ChunkIndex,
			&result.PageNumber,
			&result.Similarity,
		); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search result: %w", err)
	}

	return results, nil
}
