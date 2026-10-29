package api

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"data": data,
		})
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if message != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"error": message,
		})
	}
}
