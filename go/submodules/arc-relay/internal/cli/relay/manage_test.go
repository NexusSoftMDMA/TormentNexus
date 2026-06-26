package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupManageMock(t *testing.T, token string) (*httptest.Server, *[]http.Request) {
	t.Helper()
	var received []http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture a copy of the request
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		received = append(received, *r)

		// Auth check
		if token != "" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer "+token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		switch {
		// POST /api/servers — create
		case r.Method == http.MethodPost && r.URL.Path == "/api/servers":
			var req CreateServerRequest
			if err := json.Unmarshal(body, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
				return
			}
			if req.Name == "" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "name is required"})
				return
			}
			if req.Name == "existing-server" {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "already exists"})
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(ServerDetail{
				ID:          "new-id-123",
				Name:        req.Name,
				DisplayName: req.DisplayName,
				ServerType:  req.ServerType,
				Config:      req.Config,
				Status:      "stopped",
			})

		// DELETE /api/servers/{id}
		case r.Method == http.MethodDelete && len(r.URL.Path) > len("/api/servers/"):
			id := r.URL.Path[len("/api/servers/"):]
			if id == "nonexistent" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		// POST /api/servers/{id}/start or /stop
		case r.Method == http.MethodPost && (pathEndsWith(r.URL.Path, "/start") || pathEndsWith(r.URL.Path, "/stop")):
			w.WriteHeader(http.StatusOK)

		// GET /api/servers/{id}
		case r.Method == http.MethodGet && len(r.URL.Path) > len("/api/servers/"):
			id := r.URL.Path[len("/api/servers/"):]
			if id == "nonexistent" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(ServerDetail{
				ID:         id,
				Name:       "test-server",
				ServerType: "remote",
				Status:     "running",
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return ts, &received
}

func pathEndsWith(path, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}

func TestCreateServerRemote(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")

	cfg := RemoteConfig{
		URL:  "https://mcp.example.com/sse",
		Auth: RemoteAuth{Type: "bearer", Token: "srv-token"},
	}
	cfgJSON, _ := json.Marshal(cfg)

	detail, err := client.CreateServer(&CreateServerRequest{
		Name:        "my-server",
		DisplayName: "My MCP Server",
		ServerType:  "remote",
		Config:      cfgJSON,
	})
	if err != nil {
		t.Fatalf("CreateServer: %v", err)
	}
	if detail.ID == "" {
		t.Error("expected non-empty ID")
	}
	if detail.Name != "my-server" {
		t.Errorf("Name = %q, want %q", detail.Name, "my-server")
	}
	if detail.Status != "stopped" {
		t.Errorf("Status = %q, want %q", detail.Status, "stopped")
	}
}

func TestCreateServerStdioBuild(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")

	cfg := StdioConfig{
		Build: &StdioBuildConfig{
			Runtime: "python",
			Package: "mcp-server-sentry",
		},
	}
	cfgJSON, _ := json.Marshal(cfg)

	detail, err := client.CreateServer(&CreateServerRequest{
		Name:        "sentry",
		DisplayName: "Sentry MCP",
		ServerType:  "stdio",
		Config:      cfgJSON,
	})
	if err != nil {
		t.Fatalf("CreateServer: %v", err)
	}
	if detail.ServerType != "stdio" {
		t.Errorf("ServerType = %q, want %q", detail.ServerType, "stdio")
	}
}

func TestCreateServerConflict(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	_, err := client.CreateServer(&CreateServerRequest{
		Name:       "existing-server",
		ServerType: "remote",
		Config:     json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatal("expected error for duplicate server name")
	}
}

func TestCreateServerUnauthorized(t *testing.T) {
	ts, _ := setupManageMock(t, "correct-key")
	defer ts.Close()

	client := NewClient(ts.URL, "wrong-key")
	_, err := client.CreateServer(&CreateServerRequest{
		Name:       "test",
		ServerType: "remote",
		Config:     json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestDeleteServer(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	err := client.DeleteServer("server-id-123")
	if err != nil {
		t.Fatalf("DeleteServer: %v", err)
	}
}

func TestDeleteServerNotFound(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	err := client.DeleteServer("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent server")
	}
}

func TestStartServer(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	err := client.StartServer("server-id")
	if err != nil {
		t.Fatalf("StartServer: %v", err)
	}
}

func TestStopServer(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	err := client.StopServer("server-id")
	if err != nil {
		t.Fatalf("StopServer: %v", err)
	}
}

func TestGetServer(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	detail, err := client.GetServer("server-id-123")
	if err != nil {
		t.Fatalf("GetServer: %v", err)
	}
	if detail.ID != "server-id-123" {
		t.Errorf("ID = %q, want %q", detail.ID, "server-id-123")
	}
}

func TestGetServerNotFound(t *testing.T) {
	ts, _ := setupManageMock(t, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	_, err := client.GetServer("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent server")
	}
}
