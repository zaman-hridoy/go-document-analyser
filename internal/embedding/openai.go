package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

const openAIEmbeddedURL = "https://api.openai.com/v1/embeddings"

type OpenAI struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		apiKey: apiKey,
		model:  model,

		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type openAIEmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`

	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func (o *OpenAI) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	payload := openAIEmbeddingRequest{
		Model:          o.model,
		Input:          texts,
		EncodingFormat: "float",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	fmt.Println(body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		openAIEmbeddedURL,
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, fmt.Errorf("create embedding request")
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send enbedding request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limitedReader := io.LimitReader(resp.Body, 1<<20)
		data, _ := io.ReadAll(limitedReader)

		var apiErr openAIErrorResponse
		err := json.Unmarshal(data, &apiErr)

		if err == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("OpenAI embeddings api: %s", apiErr.Error.Message)
		}

		return nil, fmt.Errorf("OpenAI embedded api returned status: %d", resp.StatusCode)
	}

	var response openAIEmbeddingResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	if len(response.Data) != len(texts) {
		return nil, fmt.Errorf(
			"embedding count mismatch: requested %d got %d",
			len(response.Data),
			len(texts),
		)
	}

	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[j].Index < response.Data[i].Index
	})

	embeddings := make([][]float32, len(response.Data))

	for i, item := range response.Data {
		if len(item.Embedding) != 1536 {
			return nil, fmt.Errorf("unexpected embedding dimension: got %d want 1536", len(item.Embedding))
		}

		embeddings[i] = item.Embedding
	}

	return embeddings, nil
}
