package httpapi

import (
	"net/http"
)

func (s *Server) handleGetMemory(w http.ResponseWriter, r *http.Request) {
	s.handleMemoryList(w, r)
}

func (s *Server) handleExecuteCode(w http.ResponseWriter, r *http.Request) {
	s.handleCodeExec(w, r)
}

func (s *Server) handleMemorySearch(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleMemoryContexts(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleMemorySectionedStatus(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleMemoryArchiveSession(w http.ResponseWriter, r *http.Request) {}
