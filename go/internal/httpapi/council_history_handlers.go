package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

func (s *Server) handleCouncilHistoryStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.status", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.status"}})
		return
	}
	recordCount, localErr := s.debateHistory.GetRecordCount(r.Context())
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"enabled":     s.debateHistory.GetConfig().Enabled,
			"recordCount": recordCount,
			"storageSize": s.debateHistory.GetStorageSize(),
			"config":      s.debateHistory.GetConfig(),
		},
		"bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.status", "reason": "upstream unavailable; using native Go debate-history status"},
	})
}

func (s *Server) handleCouncilHistoryConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var result any
		upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.getConfig", nil, &result)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.getConfig"}})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": s.debateHistory.GetConfig(), "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.getConfig", "reason": "upstream unavailable; using native Go debate-history config"}})
	case http.MethodPost:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err.Error() != "EOF" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
			return
		}
		var result any
		upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.updateConfig", payload, &result)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.updateConfig"}})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": s.debateHistory.UpdateConfig(payload), "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.updateConfig", "reason": "upstream unavailable; updating native Go debate-history config"}})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
	}
}

func (s *Server) handleCouncilHistoryToggle(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Enabled *bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err.Error() != "EOF" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.toggle", map[string]any{"enabled": payload.Enabled}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.toggle"}})
		return
	}
	enabled := s.debateHistory.Toggle(payload.Enabled)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"enabled": enabled}, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.toggle", "reason": "upstream unavailable; toggling native Go debate-history"}})
}

func (s *Server) handleCouncilHistoryStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.stats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.stats"}})
		return
	}
	stats, localErr := s.debateHistory.GetStats(r.Context())
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": stats, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.stats", "reason": "upstream unavailable; using native Go debate-history stats"}})
}

func (s *Server) handleCouncilHistoryList(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	options := orchestration.DebateQueryOptions{Limit: 50, SortBy: "timestamp", SortOrder: "desc"}
	if sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId")); sessionID != "" {
		payload["sessionId"] = sessionID
		options.SessionID = sessionID
	}
	if taskType := strings.TrimSpace(r.URL.Query().Get("taskType")); taskType != "" {
		payload["taskType"] = taskType
		options.TaskType = taskType
	}
	if supervisorName := strings.TrimSpace(r.URL.Query().Get("supervisorName")); supervisorName != "" {
		payload["supervisorName"] = supervisorName
		options.SupervisorName = supervisorName
	}
	if approved := strings.TrimSpace(r.URL.Query().Get("approved")); approved != "" {
		if parsed, err := strconv.ParseBool(approved); err == nil {
			payload["approved"] = parsed
			options.Approved = &parsed
		}
	}
	if fromTimestamp := strings.TrimSpace(r.URL.Query().Get("fromTimestamp")); fromTimestamp != "" {
		if parsed, err := strconv.ParseInt(fromTimestamp, 10, 64); err == nil {
			payload["fromTimestamp"] = parsed
			options.FromTimestamp = &parsed
		}
	}
	if toTimestamp := strings.TrimSpace(r.URL.Query().Get("toTimestamp")); toTimestamp != "" {
		if parsed, err := strconv.ParseInt(toTimestamp, 10, 64); err == nil {
			payload["toTimestamp"] = parsed
			options.ToTimestamp = &parsed
		}
	}
	if minConsensus := strings.TrimSpace(r.URL.Query().Get("minConsensus")); minConsensus != "" {
		if parsed, err := strconv.ParseFloat(minConsensus, 64); err == nil {
			payload["minConsensus"] = parsed
			options.MinConsensus = &parsed
		}
	}
	if maxConsensus := strings.TrimSpace(r.URL.Query().Get("maxConsensus")); maxConsensus != "" {
		if parsed, err := strconv.ParseFloat(maxConsensus, 64); err == nil {
			payload["maxConsensus"] = parsed
			options.MaxConsensus = &parsed
		}
	}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
			options.Limit = parsed
		}
	}
	if offset := strings.TrimSpace(r.URL.Query().Get("offset")); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil {
			payload["offset"] = parsed
			options.Offset = parsed
		}
	}
	if sortBy := strings.TrimSpace(r.URL.Query().Get("sortBy")); sortBy != "" {
		payload["sortBy"] = sortBy
		options.SortBy = sortBy
	}
	if sortOrder := strings.TrimSpace(r.URL.Query().Get("sortOrder")); sortOrder != "" {
		payload["sortOrder"] = sortOrder
		options.SortOrder = sortOrder
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.list", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.list"}})
		return
	}
	records, total, localErr := s.debateHistory.QueryDebates(r.Context(), options)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"records": records, "meta": map[string]any{"count": len(records), "totalRecords": total}}, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.list", "reason": "upstream unavailable; using native Go debate-history records"}})
}

func (s *Server) handleCouncilHistoryGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.get", map[string]any{"id": id}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.get"}})
		return
	}
	record, localErr := s.debateHistory.GetDebate(r.Context(), id)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	if record == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "debate not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": record, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.get", "reason": "upstream unavailable; using native Go debate-history record"}})
}

func (s *Server) handleCouncilHistoryDelete(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.delete", map[string]any{"id": payload.ID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.delete"}})
		return
	}
	deleted, localErr := s.debateHistory.DeleteRecord(r.Context(), payload.ID)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"deleted": deleted, "id": payload.ID}, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.delete", "reason": "upstream unavailable; deleting native Go debate-history record"}})
}

func (s *Server) handleCouncilHistorySupervisor(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing name query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.supervisorHistory", map[string]any{"name": name}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.supervisorHistory"}})
		return
	}
	history, localErr := s.debateHistory.GetSupervisorVoteHistory(r.Context(), name)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": history, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.supervisorHistory", "reason": "upstream unavailable; using native Go supervisor vote history"}})
}

func (s *Server) handleCouncilHistoryClear(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.clear", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.clear"}})
		return
	}
	cleared, localErr := s.debateHistory.ClearAll(r.Context())
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"cleared": cleared}, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.clear", "reason": "upstream unavailable; clearing native Go debate-history"}})
}

func (s *Server) handleCouncilHistoryInitialize(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.history.initialize", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "council.history.initialize"}})
		return
	}
	recordCount, localErr := s.debateHistory.Initialize(r.Context())
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"initialized": true, "recordCount": recordCount}, "bridge": map[string]any{"fallback": "go-local-council-history", "procedure": "council.history.initialize", "reason": "upstream unavailable; initializing native Go debate-history"}})
}
