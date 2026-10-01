package rag

import (
	"context"

	"github.com/zaman-hridoy/go-document-analyser/internal/chat"
)

type QueryWriter interface {
	Rewrite(
		ctx context.Context,
		history []chat.Message,
		question string,
	) (string, error)
}
