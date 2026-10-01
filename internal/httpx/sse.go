package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteSSE(w http.ResponseWriter, event string, data any) error {
	payload, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"event: %s\ndata: %s\n\n",
		event,
		payload,
	); err != nil {
		return err
	}

	flusher, ok := w.(http.Flusher)

	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	flusher.Flush()

	return nil
}
