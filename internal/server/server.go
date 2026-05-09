package server

import (
	"net/http"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib-api/internal/handlers"
)

// New returns an HTTP server bound to listenAddr (e.g. ":8092", "127.0.0.1:8092").
// When LISTEN_ADDR is unset, main sets listenAddr to ":"+PORT.
func New(listenAddr string) *http.Server {
	mux := http.NewServeMux()
	h := handlers.New()

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/v1/parse", h.Parse)
	mux.HandleFunc("POST /api/v1/validate", h.Validate)
	mux.HandleFunc("POST /api/v1/enrich", h.Enrich)
	mux.HandleFunc("POST /api/v1/parse-validate", h.ParseValidate)
	mux.HandleFunc("POST /api/v1/parse-validate-enrich", h.Pipeline)
	mux.HandleFunc("POST /api/v1/export", h.Export)
	mux.HandleFunc("POST /api/v1/reconcile/merge-plan", h.ReconcileMergePlan)
	mux.HandleFunc("POST /api/v1/reconcile/session", h.ReconcileSession)

	return &http.Server{
		Addr:    listenAddr,
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
