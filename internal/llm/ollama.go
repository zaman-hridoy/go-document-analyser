package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Ollama struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllama(baseURL, model string) *Ollama {
	return &Ollama{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

type ollamaChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ollamaChatResponse struct {
	Message Message `json:"message"`
}

type ollamStreamResponse struct {
	Message         Message `json:"message"`
	Done            bool    `json:"done"`
	DoneReason      string  `json:"done_reason"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	EvalCount       int     `json:"eval_count"`
}

func (o *Ollama) Generate(
	ctx context.Context,
	messages []Message,
) (string, error) {
	payload := ollamaChatRequest{
		Model:    o.model,
		Messages: messages,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.baseURL+"/api/chat",
		bytes.NewReader(body),
	)

	if err != nil {
		return "", fmt.Errorf("create ollama request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send ollama request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return "", fmt.Errorf("ollama retruned status %d: %s", resp.StatusCode, string(data))
	}

	var result ollamaChatResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}

	return result.Message.Content, nil
}

func (o *Ollama) Stream(
	ctx context.Context,
	messages []Message,
	onChunk func(content string) error,
) error {
	payload := ollamaChatRequest{
		Model:    o.model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("marshal ollama stram request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.baseURL+"/api/chat",
		bytes.NewReader(body),
	)

	if err != nil {
		return fmt.Errorf("create ollama stream request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client.Do(req)

	if err != nil {
		return fmt.Errorf("send ollama stram request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

		return fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(data))
	}

	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk ollamStreamResponse
		if err := decoder.Decode(&chunk); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("decode ollama stream: %w", err)
		}

		if chunk.Message.Content != "" {
			if err := onChunk(chunk.Message.Content); err != nil {
				return fmt.Errorf("handle stream chunk: %w", err)
			}
		}

		if chunk.Done {
			break
		}
	}

	return nil

}
