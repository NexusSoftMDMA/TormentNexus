package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/mesh"
)

func (s *Server) handleMeshStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	status, err := s.mesh.Status(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": status})
}

func (s *Server) handleMeshPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	peers, err := s.mesh.Peers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": peers})
}

func (s *Server) handleMeshCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	capabilities, err := s.mesh.Capabilities(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": capabilities})
}

func (s *Server) handleMeshQueryCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	nodeID := strings.TrimSpace(r.URL.Query().Get("nodeId"))
	if nodeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing nodeId query parameter"})
		return
	}
	details, err := s.mesh.QueryCapabilities(r.Context(), nodeID, parseMeshTimeoutMs(r))
	if err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, mesh.ErrInvalidNodeID) {
			status = http.StatusBadRequest
		} else if errors.Is(err, mesh.ErrNodeNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": details})
}

func (s *Server) handleMeshFindPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	required := parseMeshCapabilityQuery(r)
	if len(required) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing capability query parameter"})
		return
	}
	match, err := s.mesh.FindPeerForCapabilities(r.Context(), required, parseMeshTimeoutMs(r))
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": match})
}

func (s *Server) handleMeshBroadcast(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "mesh.broadcast")
}

func parseMeshTimeoutMs(r *http.Request) int {
	raw := strings.TrimSpace(r.URL.Query().Get("timeoutMs"))
	if raw == "" {
		return 3000
	}
	timeoutMs, err := strconv.Atoi(raw)
	if err != nil || timeoutMs <= 0 {
		return 3000
	}
	return timeoutMs
}

func parseMeshCapabilityQuery(r *http.Request) []string {
	values := r.URL.Query()["capability"]
	if len(values) == 0 {
		return nil
	}
	capabilities := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				capabilities = append(capabilities, trimmed)
			}
		}
	}
	return capabilities
}
