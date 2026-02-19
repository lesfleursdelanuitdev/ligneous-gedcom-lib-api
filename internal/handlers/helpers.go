package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const maxUploadSize = 10 << 20

func extractGedcomFile(r *http.Request) ([]byte, error) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return nil, fmt.Errorf("parse multipart form: %w", err)
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("read file field: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file data: %w", err)
	}
	return data, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
