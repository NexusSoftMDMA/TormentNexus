package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) handleAutoDevStartLoop(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "autoDev.startLoop", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "autoDev.startLoop",
			},
		})
		return
	}

	var config localAutoDevLoopConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid request body"})
		return
	}

	loopID := s.autoDev.startLoop(config)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    loopID,
		"bridge": map[string]any{
			"fallback":  "go-local-autodev",
			"procedure": "autoDev.startLoop",
			"reason":    "upstream unavailable; starting native Go autodev loop",
		},
	})
}

func (s *Server) handleAutoDevCancelLoop(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "autoDev.cancelLoop", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "autoDev.cancelLoop",
			},
		})
		return
	}

	var payload struct {
		LoopID string `json:"loopId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid request body"})
		return
	}

	success := s.autoDev.cancelLoop(payload.LoopID)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    success,
		"bridge": map[string]any{
			"fallback":  "go-local-autodev",
			"procedure": "autoDev.cancelLoop",
			"reason":    "upstream unavailable; cancelling native Go autodev loop",
		},
	})
}

func (s *Server) handleAutoDevGetLoops(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "autoDev.getLoops", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "autoDev.getLoops",
			},
		})
		return
	}

	loops := s.autoDev.getLoops()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    loops,
		"bridge": map[string]any{
			"fallback":  "go-local-autodev",
			"procedure": "autoDev.getLoops",
			"reason":    "upstream unavailable; listing native Go autodev loops",
		},
	})
}

func (s *Server) handleAutoDevGetLoop(w http.ResponseWriter, r *http.Request) {
	loopID := strings.TrimSpace(r.URL.Query().Get("loopId"))
	if loopID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing loopId query parameter"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "autoDev.getLoop", map[string]any{"loopId": loopID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "autoDev.getLoop",
			},
		})
		return
	}

	loop, ok := s.autoDev.getLoop(loopID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "loop not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    loop,
		"bridge": map[string]any{
			"fallback":  "go-local-autodev",
			"procedure": "autoDev.getLoop",
			"reason":    "upstream unavailable; reading native Go autodev loop",
		},
	})
}

func (s *Server) handleAutoDevClearCompleted(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "autoDev.clearCompleted", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "autoDev.clearCompleted",
			},
		})
		return
	}

	count := s.autoDev.clearCompleted()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    count,
		"bridge": map[string]any{
			"fallback":  "go-local-autodev",
			"procedure": "autoDev.clearCompleted",
			"reason":    "upstream unavailable; clearing native Go autodev loops",
		},
	})
}
