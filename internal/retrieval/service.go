package retrieval

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/zaman-hridoy/go-document-analyser/internal/embedding"
)

type Service struct {
	repository *Repository
	embedder   embedding.Embedder
}

func NewService(repository *Repository, embedder embedding.Embedder) *Service {
	return &Service{
		repository: repository,
		embedder:   embedder,
	}
}

func (s *Service) Search(
	ctx context.Context,
	query string,
	documentIDs []uuid.UUID,
	limit int,
) ([]Result, error) {
	query = strings.TrimSpace(query)

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if limit <= 0 {
		limit = 5
	}

	if limit > 20 {
		limit = 20
	}

	embeddings, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	if len(embeddings) != 1 {
		return nil, fmt.Errorf("unexpected query embedding response")
	}

	results, err := s.repository.Search(ctx, embeddings[0], documentIDs, limit)
	if err != nil {
		return nil, err
	}

	return results, nil
}
