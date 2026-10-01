package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	Port        string
	DatabaseURL string
	Environment string
	UploadDir   string

	OpenAIAPIKey   string
	EmbeddingModel string

	OllamaBaseURL   string
	OllamaChatModel string

	AllowedOrigins []string
}

func MustLoad() (*config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	cfg := &config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads/documents"),

		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
		EmbeddingModel:  getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
		OllamaBaseURL:   getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaChatModel: getEnv("OLLAMA_CHAT_MODEL", "llama3.2"),
	}

	origins, ok := os.LookupEnv("ALLOWED_ORIGINS")
	if !ok {
		return nil, fmt.Errorf("ALLOWED_ORIGINS IS REQUIRED")
	}

	var allowedOrigins []string
	if origins == "" {
		return nil, fmt.Errorf("ALLOWED_ORIGINS IS REQUIRED")
	}

	for _, origin := range strings.Split(origins, ",") {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}

		allowedOrigins = append(allowedOrigins, origin)
	}

	cfg.AllowedOrigins = allowedOrigins

	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	cfg.DatabaseURL = databaseUrl

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}
