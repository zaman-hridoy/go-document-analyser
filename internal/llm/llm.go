package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Generator interface {
	Generate(
		ctx context.Context,
		messages []Message,
	) (string, error)
}

type StreamGenerator interface {
	Stream(
		ctx context.Context,
		messages []Message,
		onChunk func(content string) error,
	) error
}
