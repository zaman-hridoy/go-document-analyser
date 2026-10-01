package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repository *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repository: repo,
	}
}

func (s *Service) Create(
	ctx context.Context,
	title string,
	documentIDs []uuid.UUID,
) (*Chat, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		title = "New chat"
	}

	if len(documentIDs) == 0 {
		return nil, fmt.Errorf("at least one document is required")
	}

	return s.repository.CreateChat(ctx, title, documentIDs)
}
