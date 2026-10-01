package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ollamaEmbeddedURL = "http://localhost:11434/api/embed"

type Ollama struct {
	model     string
	dimension int
	client    *http.Client
}

func NewOllama(model string, dimension int) *Ollama {
	return &Ollama{model: model, dimension: dimension, client: &http.Client{
		Timeout: 1 * time.Minute,
	}}
}

type ollamaEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (o *Ollama) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts required")
	}

	payload := ollamaEmbeddingRequest{
		Model: o.model,
		Input: texts,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ollamaEmbeddedURL,
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, fmt.Errorf("create request embedding: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send embedding request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embedding api returned status: %d- %s", resp.StatusCode, string(data))
	}

	var response ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding embedding response: %w", err)
	}

	if len(response.Embeddings) != len(texts) {
		return nil, fmt.Errorf(
			"embedding count mismatch: requested %d got %d",
			len(texts),
			len(response.Embeddings),
		)
	}

	for i, embedding := range response.Embeddings {
		if len(embedding) != o.dimension {
			return nil, fmt.Errorf(
				"unexpected embedding dimension at index %d: got %d want %d",
				i,
				len(embedding),
				o.dimension,
			)
		}
	}

	return response.Embeddings, nil
}
