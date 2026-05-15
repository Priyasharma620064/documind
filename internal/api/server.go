package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/priya-sharma/documind/internal/search"
)

// Server implements the REST API for DocuMind.
type Server struct {
	addr   string
	engine *search.Engine
}

// NewServer creates a new API Server.
func NewServer(addr string, engine *search.Engine) *Server {
	return &Server{
		addr:   addr,
		engine: engine,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/v1/search", s.handleSearch)
	mux.HandleFunc("/api/v1/health", s.handleHealth)

	slog.Info("Starting REST API server", "addr", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	repo := r.URL.Query().Get("repo")
	version := r.URL.Query().Get("version")
	topKStr := r.URL.Query().Get("top_k")

	if query == "" || repo == "" {
		http.Error(w, "Missing 'q' or 'repo' parameter", http.StatusBadRequest)
		return
	}

	topK := 5
	if topKStr != "" {
		if val, err := strconv.Atoi(topKStr); err == nil {
			topK = val
		}
	}

	results, err := s.engine.Search(context.Background(), repo, query, topK, version)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
