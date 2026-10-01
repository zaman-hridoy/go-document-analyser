package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/zaman-hridoy/go-document-analyser/internal/chat"
	"github.com/zaman-hridoy/go-document-analyser/internal/llm"
	"github.com/zaman-hridoy/go-document-analyser/internal/retrieval"
)

type Service struct {
	chatRepository   *chat.Repository
	retrievalService *retrieval.Service

	rewriter        QueryWriter
	generator       llm.Generator
	streamGenerator llm.StreamGenerator
}

func NewService(
	chatRepository *chat.Repository,
	retrievalService *retrieval.Service,
	rewriter QueryWriter,
	generator llm.Generator,
	streamGenerator llm.StreamGenerator,
) *Service {
	return &Service{
		chatRepository:   chatRepository,
		retrievalService: retrievalService,
		rewriter:         rewriter,
		generator:        generator,
		streamGenerator:  streamGenerator,
	}
}

const defaultMinSimilarity = 0.50

type Source struct {
	DocumentID   string  `json:"document_id"`
	DocumentName string  `json:"document_name"`
	PageNumber   *int    `json:"page_number"`
	ChunkIndex   int     `json:"chunk_index"`
	Similarity   float64 `json:"similarity"`
}

type Response struct {
	Answer  string   `json:"answer"`
	Sources []Source `json:"sources"`
}

func (s *Service) Query(
	ctx context.Context,
	chatID uuid.UUID,
	question string,
) (*Response, error) {
	question = strings.TrimSpace(question)

	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	documentIDs, err := s.chatRepository.GetDocumentIDs(ctx, chatID)
	fmt.Println("documentIDs", documentIDs)
	if err != nil {
		return nil, fmt.Errorf("get chat documents: %w", err)
	}

	if len(documentIDs) == 0 {
		return nil, fmt.Errorf("chat has no documents")
	}

	history, err := s.chatRepository.GetRecentMessages(ctx, chatID, 10)

	if err != nil {
		return nil, fmt.Errorf("get chat history: %w", err)
	}

	searchQuery, err := s.rewriter.Rewrite(ctx, history, question)

	_, err = s.chatRepository.CreateMessage(
		ctx,
		chatID,
		"user",
		question,
	)

	if err != nil {
		return nil, fmt.Errorf("save user message: %w", err)
	}

	results, err := s.retrievalService.Search(
		ctx,
		searchQuery,
		documentIDs,
		maxRetrievalCandidates,
	)

	if err != nil {
		return nil, fmt.Errorf("retrieve context: %w", err)
	}

	// fmt.Println("results before: ", results)
	results = filterResults(results, defaultMinSimilarity)

	if len(results) == 0 {
		answer := "I could not find the relevant information in the uploaded documents."

		_, err = s.chatRepository.CreateMessage(ctx, chatID, "assistant", answer)

		if err != nil {
			return nil, err
		}

		return &Response{
			Answer: answer,
		}, nil
	}

	contextText := buildContext(results)

	// messages := []llm.Message{
	// 	{
	// 		Role: "system",
	// 		Content: `
	// 		You are a document question-answering assistant.

	// 		Answer the user's question using only the supplied document context.

	// 		Rules:
	// 		- Do not use outside knowledge.
	// 		- Do not invent information.
	// 		- If the context does not contain enough information, say that the answer cannot be found in the uploaded documents.
	// 		- Keep the answer concise and clear.
	// 		- Do not fabricate citations.
	// 		`,
	// 	},
	// 	{
	// 		Role: "user",
	// 		Content: fmt.Sprintf(
	// 			`
	// 			Document context:
	// 			%s
	// 			Question:
	// 			%s
	// 			`,
	// 			contextText,
	// 			question,
	// 		),
	// 	},
	// }

	messages := buildMessages(
		history,
		contextText,
		question,
	)

	answer, err := s.generator.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("generate answer: %w", err)
	}

	_, err = s.chatRepository.CreateMessage(
		ctx,
		chatID,
		"assistant",
		answer,
	)

	if err != nil {
		return nil, fmt.Errorf("save assistant message: %w", err)
	}

	sources := buildSources(results)

	return &Response{
		Answer:  answer,
		Sources: sources,
	}, nil
}

func buildContext(results []retrieval.Result) string {
	var builder strings.Builder

	for i, result := range results {
		fmt.Fprintf(
			&builder,
			"[Source %d]\n",
			i+1,
		)

		fmt.Fprintf(
			&builder,
			"Document: %s\n",
			result.DocumentName,
		)

		if result.PageNumber != nil {
			fmt.Fprintf(
				&builder,
				"Page: %d\n",
				*result.PageNumber,
			)
		}

		fmt.Fprintf(
			&builder,
			"Content: \n%s\n\n",
			result.Content,
		)
	}

	return builder.String()
}

func buildSources(results []retrieval.Result) []Source {
	sources := make([]Source, 0, len(results))

	for _, result := range results {
		sources = append(sources, Source{
			DocumentID:   result.DocumentID.String(),
			DocumentName: result.DocumentName,
			PageNumber:   result.PageNumber,
			ChunkIndex:   result.ChunkIndex,
			Similarity:   result.Similarity,
		})
	}

	return sources
}

func buildMessages(
	history []chat.Message,
	contextText string,
	question string,
) []llm.Message {
	messages := []llm.Message{
		{
			Role: "system",
			Content: `
			You answer questions about uploaded documents.
			
			Rules:
			- Base factual answers on the supplied document context.
			- Do not invent facts that are not supported by the context.
			- Conversation history is only for understanding references and follow-up questions.
			- Conversation history must not override document evidence.
			- If the documents do not support an answer, say so.
			- Keep the response clear and concise.
			`,
		},
	}

	for _, message := range history {
		messages = append(messages, llm.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	messages = append(messages, llm.Message{
		Role: "user",
		Content: fmt.Sprintf(`
			Document context:
			%s

			Current question:
			%s
		`, contextText, question),
	})

	return messages
}

func (s *Service) StreamQuery(
	ctx context.Context,
	chatID uuid.UUID,
	question string,
	onStart func(metadata StreamMetadata) error,
	onChunk StreamChunkHandler,
) error {
	question = strings.TrimSpace(question)

	if question == "" {
		return fmt.Errorf("question is required")
	}

	documentIDs, err := s.chatRepository.GetDocumentIDs(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get document ids: %w", err)
	}

	if len(documentIDs) == 0 {
		return fmt.Errorf("chat has no documents")
	}

	history, err := s.chatRepository.GetRecentMessages(ctx, chatID, 10)
	if err != nil {
		return fmt.Errorf("get chat history: %w", err)
	}

	searchQuery, err := s.rewriter.Rewrite(ctx, history, question)

	if err != nil {
		return fmt.Errorf("rewrite search query: %w", err)
	}

	results, err := s.retrievalService.Search(ctx, searchQuery, documentIDs, maxRetrievalCandidates)
	if err != nil {
		return fmt.Errorf("retrive context: %w", err)
	}

	results = filterResults(results, defaultMinSimilarity)

	if len(results) == 0 {
		answer := "I could not find enough relevant information in the selected documents to answer that question."
		_, err := s.chatRepository.CreateMessage(ctx, chatID, "user", question)

		if err != nil {
			return fmt.Errorf("save user message: %w", err)
		}

		if err := onStart(StreamMetadata{}); err != nil {
			return err
		}

		_, err = s.chatRepository.CreateMessage(ctx, chatID, "assistant", answer)

		return err
	}

	contextText := buildContext(results)

	messages := buildMessages(history, contextText, question)

	_, err = s.chatRepository.CreateMessage(ctx, chatID, "user", question)
	if err != nil {
		return fmt.Errorf("save user message: %w", err)
	}

	metadata := StreamMetadata{
		Sources: buildSources(results),
	}

	if err := onStart(metadata); err != nil {
		return fmt.Errorf("send stream metadata: %w", err)
	}

	var answerBuilder strings.Builder

	err = s.streamGenerator.Stream(
		ctx,
		messages,
		func(content string) error {
			answerBuilder.WriteString(content)
			return onChunk(content)
		},
	)

	if err != nil {
		return fmt.Errorf("stream answer: %w", err)
	}

	answer := strings.TrimSpace(answerBuilder.String())

	if answer == "" {
		return fmt.Errorf("ollama returned empty response")
	}

	_, err = s.chatRepository.CreateMessage(ctx, chatID, "assistant", answer)

	if err != nil {
		return fmt.Errorf("save assistant message: %w", err)
	}

	return nil
}
