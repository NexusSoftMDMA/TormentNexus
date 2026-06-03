package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

func (s *Server) handleAgentChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Message string       `json:"message"`
		History []ai.Message `json:"history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first for full routing/quota features
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.chat", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.chat",
			},
		})
		return
	}

	// Fall back to local Go LLM routing
	messages := payload.History
	if len(messages) == 0 && payload.Message != "" {
		messages = []ai.Message{{Role: "user", Content: payload.Message}}
	}

	llmResp, fallbackErr := ai.AutoRoute(r.Context(), messages)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"content":  llmResp.Content,
			"provider": llmResp.Provider,
			"model":    llmResp.Model,
			"usage": map[string]int{
				"inputTokens":  llmResp.Usage.InputTokens,
				"outputTokens": llmResp.Usage.OutputTokens,
			},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-llm-routing",
			"procedure": "agent.chat",
			"reason":    "upstream unavailable; using native Go LLM fallback routing",
		},
	})
}

func (s *Server) handleGoDirectorStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Goal string `json:"goal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.directorStart", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.directorStart",
			},
		})
		return
	}

	// Fallback: local Go director
	err = s.goDirector.StartAutonomousTask(r.Context(), payload.Goal)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Go autonomous director task started",
		"bridge": map[string]any{
			"fallback": "go-local-director",
		},
	})
}

func (s *Server) handleA2AListAgents(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.listA2AAgents", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.listA2AAgents",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.a2aBroker.ListAgents(),
		"bridge": map[string]any{
			"fallback":  "go-local-a2a",
			"procedure": "agent.listA2AAgents",
		},
	})
}

func (s *Server) handleA2AGetMessages(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.getA2AMessages", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.getA2AMessages",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.a2aBroker.GetHistory(),
		"bridge": map[string]any{
			"fallback":  "go-local-a2a",
			"procedure": "agent.getA2AMessages",
		},
	})
}

func (s *Server) handleA2AGetLogs(w http.ResponseWriter, r *http.Request) {
	limit := intParam(r, "limit", 100)

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.getA2ALogs", map[string]any{"limit": limit}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.getA2ALogs",
			},
		})
		return
	}

	logs, err := s.a2aLogger.GetRecentLogs(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    logs,
		"bridge": map[string]any{
			"fallback": "go-local-a2a-logs",
		},
	})
}

func (s *Server) handleA2ABroadcast(w http.ResponseWriter, r *http.Request) {
	var payload orchestration.A2AMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.a2aBroadcast", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.a2aBroadcast",
			},
		})
		return
	}

	// Fallback to local broker
	payload.Timestamp = time.Now().UnixMilli()
	if payload.ID == "" {
		payload.ID = fmt.Sprintf("a2a-%d", payload.Timestamp)
	}
	s.a2aBroker.RouteMessage(payload)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "A2A message broadcasted locally",
		"bridge": map[string]any{
			"fallback": "go-local-a2a",
		},
	})
}

func (s *Server) handleAgentSwarmStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Goal     string `json:"goal"`
		MaxTurns int    `json:"maxTurns"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.swarmStart", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.swarmStart",
			},
		})
		return
	}

	// Fallback: local Go swarm controller
	// Setup default swarm
	s.swarmController.AddMember(orchestration.SwarmMember{ID: "claude", Name: "Claude", Role: orchestration.SwarmRolePlanner, Provider: "anthropic", ModelID: "claude-3-5-sonnet-20241022", Status: "idle"})
	s.swarmController.AddMember(orchestration.SwarmMember{ID: "gpt", Name: "GPT", Role: orchestration.SwarmRoleImplementer, Provider: "openai", ModelID: "gpt-4o", Status: "idle"})
	s.swarmController.AddMember(orchestration.SwarmMember{ID: "gemini", Name: "Gemini", Role: orchestration.SwarmRoleTester, Provider: "google", ModelID: "gemini-1.5-pro", Status: "idle"})
	s.swarmController.AddMember(orchestration.SwarmMember{ID: "qwen", Name: "Qwen", Role: orchestration.SwarmRoleCritic, Provider: "google", ModelID: "gemini-2.5-flash", Status: "idle"})

	swarmResult, err := s.swarmController.StartSession(r.Context(), payload.Goal, orchestration.SwarmSessionConfig{
		MaxTurns:            payload.MaxTurns,
		CompletionThreshold: 0.8,
		AutoRotate:          true,
	})

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    swarmResult,
		"bridge": map[string]any{
			"fallback": "go-local-swarm",
		},
	})
}

func (s *Server) handleAgentSwarmTranscript(w http.ResponseWriter, r *http.Request) {
	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.getSwarmTranscript", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.getSwarmTranscript",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.swarmController.GetTranscript(),
		"bridge": map[string]any{
			"fallback": "go-local-swarm",
		},
	})
}

func (s *Server) handleAgentSupervisorEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Goal       string   `json:"goal"`
		Transcript []string `json:"transcript"`
		ModelID    string   `json:"modelId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.supervisorEvaluate", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.supervisorEvaluate",
			},
		})
		return
	}

	// Fallback: local expert supervisor
	supervisor := orchestration.NewExpertSupervisor(payload.ModelID)
	check, err := supervisor.EvaluateProgress(r.Context(), payload.Goal, payload.Transcript)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    check,
		"bridge": map[string]any{
			"fallback": "go-local-expert-supervisor",
		},
	})
}
