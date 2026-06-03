package mesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
)

func TestCapabilitiesIncludesLocalAndUpstreamNodes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trpc/mesh.getCapabilities" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"data": map[string]any{
					"json": map[string]any{
						"node-ts": []string{"git", "research"},
					},
				},
			},
		})
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	service := New(config.Default())
	capabilities, err := service.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("expected capabilities, got error: %v", err)
	}
	if _, ok := capabilities[service.LocalNodeID()]; !ok {
		t.Fatalf("expected local node in capability map, got %+v", capabilities)
	}
	got := capabilities["node-ts"]
	if len(got) != 2 || got[0] != "git" || got[1] != "research" {
		t.Fatalf("expected upstream capabilities, got %+v", got)
	}
}

func TestQueryCapabilitiesReturnsRemoteDetails(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trpc/mesh.queryCapabilities" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"data": map[string]any{
					"json": map[string]any{
						"capabilities": []string{"research", "git"},
						"role":         "typescript-main",
						"load":         0.25,
						"cachedAt":     1700000000000,
					},
				},
			},
		})
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	service := New(config.Default())
	details, err := service.QueryCapabilities(context.Background(), "node-ts", 1500)
	if err != nil {
		t.Fatalf("expected remote details, got error: %v", err)
	}
	if details.Role != "typescript-main" {
		t.Fatalf("expected role to round-trip, got %+v", details)
	}
	if details.Load == nil || *details.Load != 0.25 {
		t.Fatalf("expected load to round-trip, got %+v", details)
	}
	if len(details.Capabilities) != 2 || details.Capabilities[0] != "git" || details.Capabilities[1] != "research" {
		t.Fatalf("expected normalized capabilities, got %+v", details.Capabilities)
	}
}
