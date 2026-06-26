package relay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMockServer(t *testing.T, servers []Server, expectedToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/servers" {
			http.NotFound(w, r)
			return
		}
		if expectedToken != "" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer "+expectedToken {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servers)
	}))
}

func TestListServers(t *testing.T) {
	servers := []Server{
		{ID: "1", Name: "sentry", DisplayName: "Sentry", ServerType: "http", Status: "running"},
		{ID: "2", Name: "pfsense", DisplayName: "pfSense", ServerType: "http", Status: "running"},
		{ID: "3", Name: "broken", DisplayName: "Broken", ServerType: "http", Status: "stopped"},
	}

	ts := setupMockServer(t, servers, "test-key")
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	got, err := client.ListServers()
	if err != nil {
		t.Fatalf("ListServers: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(got))
	}
}

func TestListRunningServers(t *testing.T) {
	servers := []Server{
		{ID: "1", Name: "sentry", Status: "running"},
		{ID: "2", Name: "pfsense", Status: "running"},
		{ID: "3", Name: "broken", Status: "stopped"},
		{ID: "4", Name: "starting", Status: "starting"},
	}

	ts := setupMockServer(t, servers, "key")
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	got, err := client.ListRunningServers()
	if err != nil {
		t.Fatalf("ListRunningServers: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 running servers, got %d", len(got))
	}
	for _, s := range got {
		if s.Status != "running" {
			t.Errorf("expected running status, got %q for %s", s.Status, s.Name)
		}
	}
}

func TestListServersUnauthorized(t *testing.T) {
	ts := setupMockServer(t, nil, "correct-key")
	defer ts.Close()

	client := NewClient(ts.URL, "wrong-key")
	_, err := client.ListServers()
	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}

func TestListServersUnreachable(t *testing.T) {
	client := NewClient("http://127.0.0.1:1", "key")
	_, err := client.ListServers()
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

func TestListServersMalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	_, err := client.ListServers()
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestServerProxyURL(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "key")
	got := client.ServerProxyURL("sentry")
	want := "http://127.0.0.1:8080/mcp/sentry"
	if got != want {
		t.Errorf("ServerProxyURL = %q, want %q", got, want)
	}
}

func TestServerProxyURLTrailingSlash(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080/", "key")
	got := client.ServerProxyURL("sentry")
	want := "http://127.0.0.1:8080/mcp/sentry"
	if got != want {
		t.Errorf("ServerProxyURL = %q, want %q", got, want)
	}
}

func TestIsRelayURL(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "key")

	tests := []struct {
		url  string
		want bool
	}{
		{"http://127.0.0.1:8080/mcp/sentry", true},
		{"http://127.0.0.1:8080/mcp/pfsense", true},
		{"http://other-host:8080/mcp/sentry", false},
		{"http://127.0.0.1:8080/api/servers", false},
		{"https://example.com/mcp/sentry", false},
	}

	for _, tt := range tests {
		got := client.IsRelayURL(tt.url)
		if got != tt.want {
			t.Errorf("IsRelayURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestServerNameFromURL(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "key")

	tests := []struct {
		url  string
		want string
	}{
		{"http://127.0.0.1:8080/mcp/sentry", "sentry"},
		{"http://127.0.0.1:8080/mcp/home-assistant", "home-assistant"},
		{"http://127.0.0.1:8080/mcp/sentry/extra/path", "sentry"},
		{"http://other-host/mcp/sentry", ""},
		{"http://127.0.0.1:8080/api/servers", ""},
	}

	for _, tt := range tests {
		got := client.ServerNameFromURL(tt.url)
		if got != tt.want {
			t.Errorf("ServerNameFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestListServersHTTP500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "key")
	_, err := client.ListServers()
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}
