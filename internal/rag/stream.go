package rag

type StreamMetadata struct {
	Sources []Source `json:"sources"`
}

type StreamChunkHandler func(content string) error
