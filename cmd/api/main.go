package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zaman-hridoy/go-document-analyser/internal/chat"
	"github.com/zaman-hridoy/go-document-analyser/internal/config"
	"github.com/zaman-hridoy/go-document-analyser/internal/database"
	"github.com/zaman-hridoy/go-document-analyser/internal/document"
	"github.com/zaman-hridoy/go-document-analyser/internal/embedding"
	"github.com/zaman-hridoy/go-document-analyser/internal/httpx"
	"github.com/zaman-hridoy/go-document-analyser/internal/ingestion"
	"github.com/zaman-hridoy/go-document-analyser/internal/llm"
	"github.com/zaman-hridoy/go-document-analyser/internal/middleware"
	"github.com/zaman-hridoy/go-document-analyser/internal/rag"
	"github.com/zaman-hridoy/go-document-analyser/internal/retrieval"
)

func main() {
	ctx := context.Background()

	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg.DatabaseURL)
	cors := middleware.NewCORS(cfg.AllowedOrigins)

	// if cfg.OpenAIAPIKey == "" {
	// 	log.Fatal("OPENAI_API_KEY is required")
	// }

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	defer db.Close()

	log.Println("database connected successfully.")

	documentRepository := document.NewRepository(db)
	documentService := document.NewService(documentRepository, cfg.UploadDir)
	documentHandler := document.NewHandler(documentService)

	chunkRepository := ingestion.NewRepository(db)
	textExtractor := ingestion.NewTextExtractor()
	pdfExtractor := ingestion.NewPDFExtractor()
	chunker := ingestion.NewChunker(300, 50)
	// openAIEmbedder := embedding.NewOpenAI(cfg.OpenAIAPIKey, cfg.EmbeddingModel)
	ollamaAIEmbedder := embedding.NewOllama("embeddinggemma", 768)

	ingestionService := ingestion.NewService(
		documentRepository,
		chunkRepository,
		textExtractor,
		pdfExtractor,
		chunker,
		ollamaAIEmbedder,
	)

	retrievalRepository := retrieval.NewRepository(db)
	retrievalService := retrieval.NewService(retrievalRepository, ollamaAIEmbedder)
	retrievalHandler := retrieval.NewHandler(retrievalService)

	ingestionHandler := ingestion.NewHandler(ingestionService)

	chatRepository := chat.NewRepository(db)
	chatService := chat.NewService(chatRepository)
	chatHandler := chat.NewHandler(chatService)

	generator := llm.NewOllama(cfg.OllamaBaseURL, cfg.OllamaChatModel)
	queryWriter := rag.NewOllamaRewriter(generator)

	ragService := rag.NewService(chatRepository, retrievalService, queryWriter, generator, generator)
	ragHandler := rag.NewHandler(ragService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "OK",
		})
	})
	mux.HandleFunc("POST /api/v1/documents", documentHandler.Upload)
	mux.HandleFunc("POST /api/v1/documents/{id}/process", ingestionHandler.ProcessDocument)
	mux.HandleFunc("POST /api/v1/search", retrievalHandler.Search)
	mux.HandleFunc("POST /api/v1/query", ragHandler.Query)

	mux.HandleFunc("POST /api/v1/chats", chatHandler.Create)
	mux.HandleFunc("POST /api/v1/chats/{id}/messages", ragHandler.Query)
	mux.HandleFunc("POST /api/v1/chats/{id}/messages/stream", ragHandler.StreamQuery)

	handler := cors.Middleware(mux)

	server := http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("starting RAG API in %s Environment on port %s", cfg.Environment, cfg.Port)

		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)

	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	<-shutdownSignal

	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server stopped.")
}
