package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
		return
	}
}

func Error(
	w http.ResponseWriter,
	status int,
	message string,
) {
	WriteJSON(w, status, map[string]string{
		"error": message,
	})
}
