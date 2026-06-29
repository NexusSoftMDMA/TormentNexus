package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func (s *Server) handleMemoryAddRelation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	if s.memoryReactor == nil || s.memoryReactor.VectorStore() == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}

	var req struct {
		SourceID     string  `json:"source_id"`
		TargetID     string  `json:"target_id"`
		RelationType string  `json:"relation_type"`
		Weight       float64 `json:"weight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid json body"})
		return
	}

	if req.SourceID == "" || req.TargetID == "" || req.RelationType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "source_id, target_id, and relation_type are required"})
		return
	}

	err := s.memoryReactor.VectorStore().AddRelation(r.Context(), req.SourceID, req.TargetID, req.RelationType, req.Weight)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) handleMemoryGetRelations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	if s.memoryReactor == nil || s.memoryReactor.VectorStore() == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" && r.Method == http.MethodPost {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			id = req.ID
		}
	}

	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "id parameter is required"})
		return
	}

	relations, err := s.memoryReactor.VectorStore().GetRelations(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "relations": relations})
}

func (s *Server) handleMemorySpacedRepetitionDue(w http.ResponseWriter, r *http.Request) {
	if s.memoryReactor == nil || s.memoryReactor.VectorStore() == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if p, err := strconv.Atoi(l); err == nil && p > 0 && p <= 100 {
			limit = p
		}
	}
	allDue, err := s.memoryReactor.VectorStore().GetDueMemoriesRecords()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	if limit > 0 && len(allDue) > limit {
		allDue = allDue[:limit]
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "due_records": allDue, "total_due": len(allDue)})
}

func (s *Server) handleMemorySpacedRepetitionReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	if s.memoryReactor == nil || s.memoryReactor.VectorStore() == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}
	var req struct {
		MemoryID string `json:"memory_id"`
		Quality  int    `json:"quality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid json body"})
		return
	}
	if req.MemoryID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "memory_id is required"})
		return
	}
	err := s.memoryReactor.VectorStore().ReviewMemory(req.MemoryID, req.Quality)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}


