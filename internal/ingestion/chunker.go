package ingestion

import "strings"

type Chunk struct {
	Content    string
	ChunkIndex int
	PageNumber *int
	TokenCount int
	Embedding  []float32
}

type Chunker struct {
	chunkSize    int
	chunkOverlap int
}

func NewChunker(chunkSize int, chunkOverlap int) *Chunker {
	if chunkSize <= 0 {
		chunkSize = 300
	}

	if chunkOverlap < 0 {
		chunkOverlap = 0
	}

	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 5
	}

	return &Chunker{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

func (c Chunker) ChunkPages(
	pages []ExtractedPage,
) []Chunk {

	var chunks []Chunk

	chunkIndex := 0

	for _, page := range pages {
		text := normalizeText(page.Text)
		if text == "" {
			continue
		}

		words := strings.Fields(text)
		if len(words) == 0 {
			continue
		}

		start := 0
		for start < len(words) {
			end := start + c.chunkSize

			if end > len(words) {
				end = len(words)
			}

			chunkWords := words[start:end]
			content := strings.Join(chunkWords, " ")
			chunks = append(chunks, Chunk{
				Content:    content,
				ChunkIndex: chunkIndex,
				PageNumber: page.PageNumber,
				TokenCount: estimateToken(content),
			})

			chunkIndex++

			if end == len(words) {
				break
			}

			start = end - c.chunkOverlap
		}
	}

	return chunks
}

func normalizeText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

/*
This is only approximation
300 words might approximately 390 tokens
*/
func estimateToken(content string) int {
	words := strings.Fields(content)
	wordLength := len(words)
	if wordLength == 0 {
		return 0
	}
	return int(float64(wordLength) * 1.3)
}
