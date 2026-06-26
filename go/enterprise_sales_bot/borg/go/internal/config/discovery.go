package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ServiceDiscovery holds resolved endpoints for all TormentNexus services.
// This replaces hardcoded ports with a centralized, environment-driven configuration.
type ServiceDiscovery struct {
	// GoSidecarPort is the port the Go control plane listens on.
	GoSidecarPort int

	// TRPCUpstreamURLs are the tRPC endpoints for the TypeScript core,
	// tried in order until one responds.
	TRPCUpstreamURLs []string

	// BridgePort is the WebSocket/SSE bridge port for the TypeScript core.
	BridgePort int

	// DashboardPort is the Next.js web dashboard port.
	DashboardPort int

	// DashboardHost is the Next.js web dashboard host.
	DashboardHost string
}

// DefaultServiceDiscovery returns the standard TormentNexus service topology.
func DefaultServiceDiscovery() ServiceDiscovery {
	sd := ServiceDiscovery{
		GoSidecarPort: 4300,
		TRPCUpstreamURLs: []string{
			"http://127.0.0.1:7779/trpc",
			"http://127.0.0.1:4000/trpc",
			"http://127.0.0.1:3847/trpc",
		},
		BridgePort:    3001,
		DashboardPort: 3000,
		DashboardHost: "localhost",
	}

	// Override from environment variables
	if v := os.Getenv("TORMENTNEXUS_GO_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			sd.GoSidecarPort = p
		}
	}

	if v := strings.TrimSpace(os.Getenv("TORMENTNEXUS_TRPC_UPSTREAM")); v != "" {
		sd.TRPCUpstreamURLs = append([]string{v}, sd.TRPCUpstreamURLs...)
	}
	if v := strings.TrimSpace(os.Getenv("TORMENTNEXUS_TRPC_PORT")); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			sd.TRPCUpstreamURLs = append([]string{fmt.Sprintf("http://127.0.0.1:%d/trpc", p)}, sd.TRPCUpstreamURLs...)
		}
	}

	if v := os.Getenv("TORMENTNEXUS_BRIDGE_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			sd.BridgePort = p
		}
	}

	if v := os.Getenv("TORMENTNEXUS_DASHBOARD_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			sd.DashboardPort = p
		}
	}

	if v := os.Getenv("TORMENTNEXUS_DASHBOARD_HOST"); v != "" {
		sd.DashboardHost = v
	}

	// Deduplicate tRPC URLs
	sd.TRPCUpstreamURLs = dedupStrings(sd.TRPCUpstreamURLs)

	return sd
}

// DashboardBaseURL returns the fully qualified dashboard URL.
func (sd ServiceDiscovery) DashboardBaseURL() string {
	return "http://" + sd.DashboardHost + ":" + strconv.Itoa(sd.DashboardPort)
}

// BridgeBaseURL returns the fully qualified bridge URL.
func (sd ServiceDiscovery) BridgeBaseURL() string {
	return "http://127.0.0.1:" + strconv.Itoa(sd.BridgePort)
}

// GoSidecarBaseURL returns the fully qualified Go sidecar URL.
func (sd ServiceDiscovery) GoSidecarBaseURL() string {
	return "http://127.0.0.1:" + strconv.Itoa(sd.GoSidecarPort)
}

func dedupStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(strings.TrimRight(item, "/"))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
