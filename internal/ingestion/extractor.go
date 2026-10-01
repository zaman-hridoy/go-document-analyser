package ingestion

import "context"

type ExtractedPage struct {
	PageNumber *int
	Text       string
}

type Extractor interface {
	Extract(ctx context.Context, filePath string) ([]ExtractedPage, error)
}
