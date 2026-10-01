package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/zaman-hridoy/go-document-analyser/internal/chat"
	"github.com/zaman-hridoy/go-document-analyser/internal/llm"
)

type OllamaRewriter struct {
	generator llm.Generator
}

func NewOllamaRewriter(generator llm.Generator) *OllamaRewriter {
	return &OllamaRewriter{
		generator: generator,
	}
}

func (r *OllamaRewriter) Rewrite(
	ctx context.Context,
	history []chat.Message,
	question string,
) (string, error) {
	question = strings.TrimSpace(question)

	if question == "" {
		return "", fmt.Errorf("question is required")
	}

	if len(history) == 0 {
		return question, nil
	}

	historyText := buildRewrireHistory(history)

	messages := []llm.Message{
		{
			Role: "system",
			Content: `
			You rewrite follow-up questions into standalone search queries.

			Rules:
			- Preserve the user's original meaning.
			- Resolve pronouns and references using the conversation history.
			- Do not answer the question.
			- Do not add facts.
			- Return only the rewritten query.
			- Keep it concise.
			`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(
				`
				Conversation:
				%s

				Current question:
				%s

				Rewrite the current question as a standalone search query.
				`,
				historyText,
				question,
			),
		},
	}

	result, err := r.generator.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("rewrite query: %w", err)
	}

	result = strings.TrimSpace(result)

	if result == "" {
		return question, nil
	}

	return result, nil
}

func buildRewrireHistory(history []chat.Message) string {
	var builder strings.Builder

	for _, message := range history {
		switch message.Role {
		case "user":
			builder.WriteString("User: ")
		case "assistant":
			builder.WriteString("Assistant")
		default:
			continue
		}

		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}

	return builder.String()
}
