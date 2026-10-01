package document

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Created(
	ctx context.Context,
	doc *Document,
) error {
	const query = `
		INSERT INTO documents (
			id,
			original_name,
			stored_name,
			mime_type,
			file_size,
			storage_path,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		RETURNING
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		doc.ID,
		doc.OriginalName,
		doc.StoredName,
		doc.MimeType,
		doc.FileSize,
		doc.StoragePath,
		doc.Status,
	).Scan(
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Document, error) {
	const query = `
		SELECT
			id,
			original_name,
			stored_name,
			mime_type,
			file_size,
			storage_path,
			status,
			error_message,
			created_at,
			updated_at
		FROM documents
		WHERE id = $1
	`

	var doc Document

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.OriginalName,
		&doc.StoredName,
		&doc.MimeType,
		&doc.FileSize,
		&doc.StoragePath,
		&doc.Status,
		&doc.ErrorMessage,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	return &doc, nil
}

func (r Repository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status Status,
	errorMessage *string,
) error {
	const query = `
		UPDATE documents
		SET 
			status = $1,
			error_message = $2,
			updated_at = NOW()
		WHERE id = $3
	`

	commandTag, err := r.db.Exec(
		ctx,
		query,
		status,
		errorMessage,
		id,
	)

	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}
