package ingestion

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/giraffesyo/pdf"
)

type PDFExtractor struct{}

func NewPDFExtractor() *PDFExtractor {
	return &PDFExtractor{}
}

func (e *PDFExtractor) Extract(
	ctx context.Context,
	filePath string,
) ([]ExtractedPage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open PDF: %w", err)
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat PDF: %w", err)
	}

	doc, err := pdf.Extract(ctx, file, info.Size())
	if err != nil {
		return nil, fmt.Errorf("extract pdf: %w", err)
	}

	pages := make([]ExtractedPage, 0, len(doc.Pages))

	for i, page := range doc.Pages {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		text := strings.TrimSpace(page.Text())

		if text == "" {
			continue
		}

		pageNumber := i + 1

		pages = append(pages, ExtractedPage{
			PageNumber: &pageNumber,
			Text:       text,
		})
	}

	return pages, nil
}
