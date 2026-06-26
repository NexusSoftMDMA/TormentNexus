package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func (s *Server) handleDarwinEvolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "darwin.evolve", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "darwin.evolve",
			},
		})
		return
	}

	prompt, _ := payload["prompt"].(string)
	goal, _ := payload["goal"].(string)
	if strings.TrimSpace(prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing prompt"})
		return
	}
	mutation := s.darwinState.proposeMutation(prompt, goal)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"mutationId":    mutation.ID,
			"mutation":      mutation,
			"mutatedPrompt": mutation.MutatedPrompt,
			"reasoning":     mutation.Reasoning,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-darwin",
			"procedure": "darwin.evolve",
			"reason":    "upstream unavailable; using native Go Darwin fallback mutation scaffold",
		},
	})
}

func (s *Server) handleDarwinExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "darwin.experiment", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "darwin.experiment",
			},
		})
		return
	}

	mutationID, _ := payload["mutationId"].(string)
	task, _ := payload["task"].(string)
	if strings.TrimSpace(mutationID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing mutationId"})
		return
	}
	experiment, ok := s.darwinState.startExperiment(mutationID, task)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": errors.New("mutation not found").Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"experimentId": experiment.ID,
			"experiment":   experiment,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-darwin",
			"procedure": "darwin.experiment",
			"reason":    "upstream unavailable; using native Go Darwin fallback experiment state",
		},
	})
}

func (s *Server) handleDarwinStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "darwin.getStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "darwin.getStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.darwinState.status(),
		"bridge": map[string]any{
			"fallback":  "go-local-darwin",
			"procedure": "darwin.getStatus",
			"reason":    "upstream unavailable; using native Go Darwin fallback status",
		},
	})
}
