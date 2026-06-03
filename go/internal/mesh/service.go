package mesh

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
	"github.com/tormentnexushq/tormentnexus-go/internal/interop"
)

var (
	ErrInvalidNodeID = errors.New("missing nodeId")
	ErrNodeNotFound  = errors.New("mesh node not found")
)

type Status struct {
	NodeID     string `json:"nodeId"`
	PeersCount int    `json:"peersCount"`
}

type CapabilityDetails struct {
	Capabilities []string `json:"capabilities"`
	Role         string   `json:"role,omitempty"`
	Load         *float64 `json:"load,omitempty"`
	CachedAt     int64    `json:"cachedAt"`
}

type MatchingPeer struct {
	NodeID       string   `json:"nodeId"`
	Capabilities []string `json:"capabilities"`
	Role         string   `json:"role,omitempty"`
	Load         *float64 `json:"load,omitempty"`
}

type Service struct {
	cfg         config.Config
	localNodeID string
}

func New(cfg config.Config) *Service {
	return &Service{
		cfg:         cfg,
		localNodeID: deriveLocalNodeID(cfg),
	}
}

func (s *Service) LocalNodeID() string {
	return s.localNodeID
}

func (s *Service) Status(ctx context.Context) (Status, error) {
	peers, err := s.Peers(ctx)
	if err != nil {
		return Status{}, err
	}

	return Status{
		NodeID:     s.localNodeID,
		PeersCount: len(peers),
	}, nil
}

func (s *Service) Peers(ctx context.Context) ([]string, error) {
	capabilities, err := s.Capabilities(ctx)
	if err != nil {
		return nil, err
	}

	peers := make([]string, 0, len(capabilities))
	for nodeID := range capabilities {
		if nodeID == s.localNodeID {
			continue
		}
		peers = append(peers, nodeID)
	}
	sort.Strings(peers)
	return peers, nil
}

func (s *Service) Capabilities(ctx context.Context) (map[string][]string, error) {
	combined := map[string][]string{
		s.localNodeID: append([]string(nil), s.localCapabilities()...),
	}

	upstream, err := s.fetchUpstreamCapabilities(ctx)
	if err == nil {
		for nodeID, capabilities := range upstream {
			combined[nodeID] = capabilities
		}
	}

	return combined, nil
}

func (s *Service) QueryCapabilities(ctx context.Context, nodeID string, timeoutMs int) (CapabilityDetails, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return CapabilityDetails{}, ErrInvalidNodeID
	}

	if nodeID == s.localNodeID {
		return CapabilityDetails{
			Capabilities: append([]string(nil), s.localCapabilities()...),
			Role:         "go-sidecar",
			CachedAt:     time.Now().UnixMilli(),
		}, nil
	}

	if timeoutMs <= 0 {
		timeoutMs = 3000
	}

	upstreamResult, err := interop.CallTRPCProcedure(ctx, s.cfg.MainLockPath(), "mesh.queryCapabilities", map[string]any{
		"nodeId":    nodeID,
		"timeoutMs": timeoutMs,
	})
	if err == nil {
		var details CapabilityDetails
		if unmarshalErr := json.Unmarshal(upstreamResult.Data, &details); unmarshalErr == nil {
			if details.CachedAt == 0 {
				details.CachedAt = time.Now().UnixMilli()
			}
			details.Capabilities = normalizeCapabilities(details.Capabilities)
			return details, nil
		}
	}

	capabilities, capErr := s.Capabilities(ctx)
	if capErr != nil {
		return CapabilityDetails{}, capErr
	}

	if cached, ok := capabilities[nodeID]; ok {
		return CapabilityDetails{
			Capabilities: append([]string(nil), cached...),
			CachedAt:     time.Now().UnixMilli(),
		}, nil
	}

	return CapabilityDetails{}, ErrNodeNotFound
}

func (s *Service) FindPeerForCapabilities(ctx context.Context, required []string, timeoutMs int) (*MatchingPeer, error) {
	required = normalizeCapabilities(required)
	if len(required) == 0 {
		return nil, nil
	}

	peers, err := s.Peers(ctx)
	if err != nil {
		return nil, err
	}

	for _, nodeID := range peers {
		details, detailsErr := s.QueryCapabilities(ctx, nodeID, timeoutMs)
		if detailsErr != nil {
			continue
		}
		if !containsAll(details.Capabilities, required) {
			continue
		}

		return &MatchingPeer{
			NodeID:       nodeID,
			Capabilities: details.Capabilities,
			Role:         details.Role,
			Load:         details.Load,
		}, nil
	}

	return nil, nil
}

func (s *Service) fetchUpstreamCapabilities(ctx context.Context) (map[string][]string, error) {
	result, err := interop.CallTRPCProcedure(ctx, s.cfg.MainLockPath(), "mesh.getCapabilities", nil)
	if err != nil {
		return nil, err
	}

	var capabilities map[string][]string
	if err := json.Unmarshal(result.Data, &capabilities); err != nil {
		return nil, err
	}
	for nodeID, values := range capabilities {
		capabilities[nodeID] = normalizeCapabilities(values)
	}
	return capabilities, nil
}

func (s *Service) localCapabilities() []string {
	return []string{
		"cli-harnesses",
		"http-bridge",
		"memory-status",
		"provider-status",
		"runtime-status",
		"session-import",
	}
}

func deriveLocalNodeID(cfg config.Config) string {
	sum := sha1.Sum([]byte(cfg.ConfigDir + "|" + cfg.WorkspaceRoot))
	return "tormentnexus-go-" + hex.EncodeToString(sum[:6])
}

func normalizeCapabilities(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	sort.Strings(normalized)
	return normalized
}

func containsAll(capabilities []string, required []string) bool {
	available := make(map[string]struct{}, len(capabilities))
	for _, capability := range capabilities {
		available[capability] = struct{}{}
	}
	for _, capability := range required {
		if _, ok := available[capability]; !ok {
			return false
		}
	}
	return true
}
