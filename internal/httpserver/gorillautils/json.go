package gorillautils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RenderJSON renders value as JSON in the HTTP response.
func RenderJSON(w http.ResponseWriter, value interface{}) error {
	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(value); err != nil {
		return fmt.Errorf("failed to write json response: %w", err)
	}
	return nil
}
