package httpapi

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleSquadList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "squad.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "squad.list",
			},
		})
		return
	}

	members := s.squad.List()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    members,
		"bridge": map[string]any{
			"fallback":  "go-local-squad",
			"procedure": "squad.list",
			"reason":    "upstream unavailable; listing native Go squad members",
		},
	})
}

func (s *Server) handleSquadSpawn(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "squad.spawn", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "squad.spawn",
			},
		})
		return
	}

	var payload struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid request body"})
		return
	}

	memberId := s.squad.Spawn(payload.Role)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"memberId": memberId,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-squad",
			"procedure": "squad.spawn",
			"reason":    "upstream unavailable; spawning native Go squad member",
		},
	})
}

func (s *Server) handleSquadKill(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "squad.kill", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "squad.kill",
			},
		})
		return
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid request body"})
		return
	}

	success := s.squad.Kill(payload.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    success,
		"bridge": map[string]any{
			"fallback":  "go-local-squad",
			"procedure": "squad.kill",
			"reason":    "upstream unavailable; killing native Go squad member",
		},
	})
}

func (s *Server) handleSquadChat(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "squad.chat")
}

func (s *Server) handleSquadToggleIndexer(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "squad.toggleIndexer", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "squad.toggleIndexer",
			},
		})
		return
	}

	var payload struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid request body"})
		return
	}

	s.squad.ToggleIndexer(payload.Active)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    true,
		"bridge": map[string]any{
			"fallback":  "go-local-squad",
			"procedure": "squad.toggleIndexer",
			"reason":    "upstream unavailable; toggling native Go squad indexer",
		},
	})
}

func (s *Server) handleSquadIndexerStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "squad.getIndexerStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "squad.getIndexerStatus",
			},
		})
		return
	}

	active := s.squad.GetIndexerStatus()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"active": active,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-squad",
			"procedure": "squad.getIndexerStatus",
			"reason":    "upstream unavailable; reading native Go squad indexer status",
		},
	})
}
