package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// Server mirrors the relay API server response shape.
type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	ServerType  string `json:"server_type"`
	Status      string `json:"status"`
}

// NewMockRelay returns an httptest.Server that responds to GET /api/servers
// with the provided server list. It validates the Bearer token if expectedToken
// is non-empty.
func NewMockRelay(servers []Server, expectedToken string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/servers" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if expectedToken != "" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer "+expectedToken {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servers)
	}))
}
