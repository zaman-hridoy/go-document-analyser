package middleware

import (
	"net/http"
	"strings"

	"github.com/zaman-hridoy/go-document-analyser/internal/httpx"
)

type CORS struct {
	allowedOrigins map[string]struct{}
}

func NewCORS(allowedOrigins []string) *CORS {
	origins := make(map[string]struct{}, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)

		if origin == "" {
			continue
		}

		origins[origin] = struct{}{}
	}

	return &CORS{
		allowedOrigins: origins,
	}
}

func (c *CORS) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if _, ok := c.allowedOrigins[origin]; !ok {
			httpx.WriteJSON(w, http.StatusForbidden, map[string]string{
				"error": "origin not allowed",
			})
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			c.handlePreflight(w)
			return
		}

		next.ServeHTTP(w, r)

	})
}

func (c *CORS) handlePreflight(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
}
