package config

import (
	"testing"
)

func TestDefaultServiceDiscovery(t *testing.T) {
	sd := DefaultServiceDiscovery()
	if sd.GoSidecarPort != 4300 {
		t.Errorf("expected GoSidecarPort=4300, got %d", sd.GoSidecarPort)
	}
	if sd.BridgePort != 3001 {
		t.Errorf("expected BridgePort=3001, got %d", sd.BridgePort)
	}
	if sd.DashboardPort != 3000 {
		t.Errorf("expected DashboardPort=3000, got %d", sd.DashboardPort)
	}
	if len(sd.TRPCUpstreamURLs) != 3 {
		t.Errorf("expected 3 default tRPC URLs, got %d", len(sd.TRPCUpstreamURLs))
	}
}

func TestServiceDiscoveryFromEnv(t *testing.T) {
	t.Setenv("TORMENTNEXUS_GO_PORT", "5500")
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://192.168.1.100:7779/trpc")
	t.Setenv("TORMENTNEXUS_BRIDGE_PORT", "4001")
	t.Setenv("TORMENTNEXUS_DASHBOARD_PORT", "8080")

	sd := DefaultServiceDiscovery()

	if sd.GoSidecarPort != 5500 {
		t.Errorf("expected GoSidecarPort=5500, got %d", sd.GoSidecarPort)
	}
	if sd.BridgePort != 4001 {
		t.Errorf("expected BridgePort=4001, got %d", sd.BridgePort)
	}
	if sd.DashboardPort != 8080 {
		t.Errorf("expected DashboardPort=8080, got %d", sd.DashboardPort)
	}
	if sd.TRPCUpstreamURLs[0] != "http://192.168.1.100:7779/trpc" {
		t.Errorf("expected env tRPC URL first, got %s", sd.TRPCUpstreamURLs[0])
	}
}

func TestServiceDiscoveryDedupTRPCURLs(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:7779/trpc")
	sd := DefaultServiceDiscovery()

	// The env URL should be first but not duplicated
	count := 0
	for _, u := range sd.TRPCUpstreamURLs {
		if u == "http://127.0.0.1:7779/trpc" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of http://127.0.0.1:7779/trpc, got %d", count)
	}
}

func TestServiceDiscoveryBaseURLs(t *testing.T) {
	sd := ServiceDiscovery{
		DashboardHost:  "myhost",
		DashboardPort:  3000,
		BridgePort:     3001,
		GoSidecarPort:  4300,
	}

	if sd.DashboardBaseURL() != "http://myhost:3000" {
		t.Errorf("expected http://myhost:3000, got %s", sd.DashboardBaseURL())
	}
	if sd.BridgeBaseURL() != "http://127.0.0.1:3001" {
		t.Errorf("expected http://127.0.0.1:3001, got %s", sd.BridgeBaseURL())
	}
	if sd.GoSidecarBaseURL() != "http://127.0.0.1:4300" {
		t.Errorf("expected http://127.0.0.1:4300, got %s", sd.GoSidecarBaseURL())
	}
}

func TestServiceDiscoveryInvalidEnvPort(t *testing.T) {
	t.Setenv("TORMENTNEXUS_GO_PORT", "not-a-number")
	sd := DefaultServiceDiscovery()

	// Should fall back to default
	if sd.GoSidecarPort != 4300 {
		t.Errorf("expected default GoSidecarPort=4300, got %d", sd.GoSidecarPort)
	}
}

func TestDedupStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty", []string{}, []string{}},
		{"no dups", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"with dups", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"trailing slash normalization", []string{"http://x:7779/trpc/", "http://x:7779/trpc"}, []string{"http://x:7779/trpc"}},
		{"whitespace trim", []string{"  a  ", "a"}, []string{"a"}},
		{"empty strings skipped", []string{"", "a", ""}, []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dedupStrings(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}
