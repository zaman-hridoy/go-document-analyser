package ingestion

import (
	"context"
	"fmt"
	"os"
)

type TextExtractor struct{}

func NewTextExtractor() *TextExtractor {
	return &TextExtractor{}
}

func (e *TextExtractor) Extract(
	ctx context.Context,
	filePath string,
) ([]ExtractedPage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read text file: %w", err)
	}

	extractedPage := ExtractedPage{
		PageNumber: nil,
		Text:       string(data),
	}

	return []ExtractedPage{
		extractedPage,
	}, nil
}
