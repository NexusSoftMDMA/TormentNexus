package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) handleMemoryList(w http.ResponseWriter, r *http.Request) {
	memories := s.memoryManager.GetMemories()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memories)
}

func (s *Server) handleMemoryAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.memoryManager.AddMemory(req.Content)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleMemoryAddHistory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		History []map[string]any `json:"history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, item := range req.History {
		content := fmt.Sprintf("Visited: %v (%v)", item["title"], item["url"])
		s.memoryManager.AddMemory(content)
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "count": len(req.History)})
}
