package chat

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(
	db *pgxpool.Pool,
) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateChat(
	ctx context.Context,
	title string,
	documentIDs []uuid.UUID,
) (*Chat, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create chat transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	var chat Chat
	err = tx.QueryRow(
		ctx,
		`
			INSERT INTO chats (title)
			VALUES ($1)
			RETURNING
				id,
				title,
				created_at,
				updated_at
		`,
		title,
	).Scan(
		&chat.ID,
		&chat.Title,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("creat chat: %w", err)
	}

	for _, documentID := range documentIDs {
		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO chat_documents (
				chat_id,
				document_id
			)
			VALUES ($1, $2)
			`,
			chat.ID,
			documentID,
		)

		if err != nil {
			return nil, fmt.Errorf("attach document to chat: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create chat: %w", err)
	}

	return &chat, nil
}

func (r Repository) CreateMessage(
	ctx context.Context,
	chatID uuid.UUID,
	role string,
	content string,
) (*Message, error) {
	var message Message
	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO chat_messages (
			chat_id,
			role,
			content
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			chat_id,
			role,
			content,
			created_at
		`,
		chatID,
		role,
		content,
	).Scan(
		&message.ID,
		&message.ChatID,
		&message.Role,
		&message.Content,
		&message.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create chat message: %w", err)
	}

	return &message, nil
}

func (r *Repository) GetRecentMessages(
	ctx context.Context,
	chatID uuid.UUID,
	limit int,
) ([]Message, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			chat_id,
			role,
			content,
			created_at
		FROM (
			SELECT
				id,
				chat_id,
				role,
				content,
				created_at
			FROM chat_messages
			WHERE chat_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		) recent
		ORDER BY created_at ASC, id ASC
		`,
		chatID,
		limit,
	)

	if err != nil {
		return nil, fmt.Errorf("get recent messages: %w", err)
	}

	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		if err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.Role,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat messages: %w", err)
	}

	return messages, nil
}

func (r *Repository) GetDocumentIDs(
	ctx context.Context,
	chatID uuid.UUID,
) ([]uuid.UUID, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT document_id
		FROM chat_documents
		WHERE chat_id = $1
		`,
		chatID,
	)

	if err != nil {
		return nil, fmt.Errorf("get chat documents: %w", err)
	}

	defer rows.Close()

	var documents []uuid.UUID
	for rows.Next() {
		var docID uuid.UUID
		if err := rows.Scan(&docID); err != nil {
			return nil, fmt.Errorf("scan doc id: %w", err)
		}
		documents = append(documents, docID)
	}

	return documents, rows.Err()
}
