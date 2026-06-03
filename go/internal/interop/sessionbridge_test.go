package interop

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/tormentnexushq/tormentnexus-go/internal/lockfile"
)

func TestResolveTRPCBasesUsesConfiguredExclusivelyWhenSet(t *testing.T) {
	tempDir := t.TempDir()
	mainLockPath := filepath.Join(tempDir, "lock")
	// Seed a lock file — this should be IGNORED when TORMENTNEXUS_TRPC_UPSTREAM is set
	if err := lockfile.Write(mainLockPath, lockfile.Record{
		Host: "0.0.0.0",
		Port: 4100,
	}); err != nil {
		t.Fatalf("failed to seed lock file: %v", err)
	}
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:4200/trpc")
	bases := ResolveTRPCBases(mainLockPath)
	if len(bases) != 1 {
		t.Fatalf("expected exactly 1 base when TORMENTNEXUS_TRPC_UPSTREAM is set, got %v", bases)
	}
	if bases[0] != "http://127.0.0.1:4200/trpc" {
		t.Fatalf("expected configured base, got %s", bases[0])
	}
}

func TestResolveTRPCBasesUsesLockfileAndDefaultsWhenNoEnv(t *testing.T) {
	tempDir := t.TempDir()
	mainLockPath := filepath.Join(tempDir, "lock")
	if err := lockfile.Write(mainLockPath, lockfile.Record{
		Host: "0.0.0.0",
		Port: 4100,
	}); err != nil {
		t.Fatalf("failed to seed lock file: %v", err)
	}
	// Ensure TORMENTNEXUS_TRPC_UPSTREAM is not set
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "")
	bases := ResolveTRPCBases(mainLockPath)
	if len(bases) < 2 {
		t.Fatalf("expected locked + default bases, got %v", bases)
	}
	if bases[0] != "http://127.0.0.1:4100/trpc" {
		t.Fatalf("expected locked base first, got %s", bases[0])
	}
}

func TestCallTRPCProcedureReturnsUnwrappedJSONData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trpc/session.list" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"data": map[string]any{
					"json": []map[string]any{
						{"id": "sess_1", "status": "running"},
					},
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", server.URL+"/trpc")
	result, err := CallTRPCProcedure(context.Background(), filepath.Join(t.TempDir(), "missing-lock"), "session.list", nil)
	if err != nil {
		t.Fatalf("expected no bridge error, got %v", err)
	}
	if result.BaseURL != server.URL+"/trpc" {
		t.Fatalf("expected test server base url, got %q", result.BaseURL)
	}
	if string(result.Data) != `[{"id":"sess_1","status":"running"}]` {
		t.Fatalf("expected unwrapped tRPC payload, got %s", string(result.Data))
	}
}
