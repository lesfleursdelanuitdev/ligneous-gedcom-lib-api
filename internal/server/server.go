package server

import (
	"fmt"
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib-api/internal/handlers"
)

func New(port int) *http.Server {
	mux := http.NewServeMux()
	h := handlers.New()

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/v1/parse", h.Parse)
	mux.HandleFunc("POST /api/v1/validate", h.Validate)
	mux.HandleFunc("POST /api/v1/enrich", h.Enrich)
	mux.HandleFunc("POST /api/v1/parse-validate", h.ParseValidate)
	mux.HandleFunc("POST /api/v1/parse-validate-enrich", h.Pipeline)
	mux.HandleFunc("POST /api/v1/export", h.Export)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: corsMiddleware(mux),
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
