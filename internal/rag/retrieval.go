package rag

import (
	"fmt"

	"github.com/zaman-hridoy/go-document-analyser/internal/retrieval"
)

const (
	maxRetrievalCandidates = 12
	maxContextChunks       = 5
)

func filterResults(
	results []retrieval.Result,
	minSimilarity float64,
) []retrieval.Result {
	filtered := make([]retrieval.Result, 0, maxContextChunks)

	seen := make(map[string]struct{})

	for _, result := range results {
		fmt.Println("Similarity", result.Similarity)
		if result.Similarity < minSimilarity {
			continue
		}

		key := result.DocumentID.String()

		if result.PageNumber != nil {
			key += ":" + fmt.Sprintf("%d", *result.PageNumber)
		} else {
			key += ":" + fmt.Sprintf("%d", result.ChunkIndex)
		}

		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		filtered = append(filtered, result)

		if len(filtered) >= maxContextChunks {
			break
		}
	}

	return filtered
}
