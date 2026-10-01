package ingestion

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/zaman-hridoy/go-document-analyser/internal/document"
	"github.com/zaman-hridoy/go-document-analyser/internal/embedding"
)

type Service struct {
	documentRepository *document.Repository
	chunkRepository    *Repository

	textExtractor *TextExtractor
	pdfExtractor  *PDFExtractor

	chunker  *Chunker
	embedder embedding.Embedder
}

func NewService(
	documentRepository *document.Repository,
	chunkRepository *Repository,
	textExtractor *TextExtractor,
	pdfExtractor *PDFExtractor,
	chunker *Chunker,
	embedder embedding.Embedder,
) *Service {
	return &Service{
		documentRepository: documentRepository,
		chunkRepository:    chunkRepository,
		textExtractor:      textExtractor,
		chunker:            chunker,
		embedder:           embedder,
	}
}

func (s *Service) Process(
	ctx context.Context,
	documentID uuid.UUID,
) error {

	doc, err := s.documentRepository.GetByID(ctx, documentID)

	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}

	err = s.documentRepository.UpdateStatus(ctx, documentID, document.StatusProcessing, nil)
	if err != nil {
		return err
	}

	pages, err := s.extract(ctx, doc)

	if err != nil {
		s.markFailed(ctx, documentID, err.Error())
		return err
	}

	chunks := s.chunker.ChunkPages(pages)
	if len(chunks) == 0 {
		err := fmt.Errorf("no extractable text found")
		s.markFailed(ctx, documentID, err.Error())
		return err
	}

	if err := s.embedChunks(ctx, chunks); err != nil {
		s.markFailed(ctx, documentID, err.Error())
		return fmt.Errorf("embed chunks: %w", err)
	}

	err = s.chunkRepository.ReplaceChunks(
		ctx,
		documentID,
		chunks,
	)

	if err != nil {
		s.markFailed(ctx, documentID, err.Error())
		return err
	}

	err = s.documentRepository.UpdateStatus(ctx, documentID, document.StatusReady, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) extract(
	ctx context.Context,
	doc *document.Document,
) ([]ExtractedPage, error) {

	switch {
	case doc.MimeType == "application/pdf":
		return s.pdfExtractor.Extract(ctx, doc.StoragePath)
	case strings.HasPrefix(doc.MimeType, "text/plain"):
		return s.textExtractor.Extract(ctx, doc.StoragePath)
	default:
		return nil, fmt.Errorf("unsupported mime type for ingestion: %s", doc.MimeType)
	}
}

func (s *Service) markFailed(
	ctx context.Context,
	documentID uuid.UUID,
	message string,
) {
	_ = s.documentRepository.UpdateStatus(
		ctx,
		documentID,
		document.StatusFailed,
		&message,
	)
}

const embeddingBatchSize = 32

func (s *Service) embedChunks(
	ctx context.Context,
	chunks []Chunk,
) error {
	for start := 0; start < len(chunks); start += embeddingBatchSize {
		end := start + embeddingBatchSize

		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[start:end]

		texts := make([]string, len(batch))

		for i, chunk := range batch {
			texts[i] = chunk.Content
		}

		embeddings, err := s.embedder.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("embedding batch %d-%d:%w", start, end-1, err)
		}

		if len(embeddings) != len(batch) {
			return fmt.Errorf("embedding result count mismatch")
		}

		for i := range batch {
			chunks[start+i].Embedding = embeddings[i]
		}
	}

	return nil
}
