package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *Server) localMemoryContextsPath() string {
	return filepath.Join(s.cfg.WorkspaceRoot, ".tormentnexus", "memory", "contexts.json")
}

func (s *Server) localWriteMemoryContexts(contexts []map[string]any) error {
	path := s.localMemoryContextsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(contexts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func localMemoryContextID(context map[string]any, fallbackIndex int) string {
	id := strings.TrimSpace(stringValue(context["id"]))
	if id != "" {
		return id
	}
	uuid := strings.TrimSpace(stringValue(context["uuid"]))
	if uuid != "" {
		return uuid
	}
	if fallbackIndex > 0 {
		return "context-" + strconv.Itoa(fallbackIndex)
	}
	return ""
}

func localMemoryContextCreatedAt(context map[string]any) string {
	raw := context["createdAt"]
	switch value := raw.(type) {
	case string:
		if strings.TrimSpace(value) != "" {
			return value
		}
	case float64:
		return time.UnixMilli(int64(value)).UTC().Format(time.RFC3339)
	case int64:
		return time.UnixMilli(value).UTC().Format(time.RFC3339)
	case int:
		return time.UnixMilli(int64(value)).UTC().Format(time.RFC3339)
	}
	return time.Now().UTC().Format(time.RFC3339)
}

func localMemorySearchDocument(record map[string]any) string {
	parts := []string{
		stringValue(record["content"]),
		stringValue(record["title"]),
		stringValue(record["source"]),
		stringValue(record["url"]),
	}
	if metadata, ok := record["metadata"].(map[string]any); ok {
		if raw, err := json.Marshal(metadata); err == nil {
			parts = append(parts, string(raw))
		}
	}
	return strings.Join(localUniqueStrings(nil, parts...), " ")
}

func (s *Server) localFindMemoryContext(id string) (map[string]any, bool, error) {
	contexts, err := s.localMemoryContexts()
	if err != nil {
		return nil, false, err
	}
	target := strings.TrimSpace(id)
	if target == "" {
		return nil, false, nil
	}
	for index, context := range contexts {
		candidateID := localMemoryContextID(context, index+1)
		if candidateID == target {
			copy := cloneMap(context)
			if _, exists := copy["id"]; !exists {
				copy["id"] = candidateID
			}
			return copy, true, nil
		}
	}
	return nil, false, nil
}

func (s *Server) localDeleteMemoryContext(id string) (bool, error) {
	contexts, err := s.localMemoryContexts()
	if err != nil {
		return false, err
	}
	target := strings.TrimSpace(id)
	if target == "" {
		return false, nil
	}
	updated := make([]map[string]any, 0, len(contexts))
	deleted := false
	for index, context := range contexts {
		if localMemoryContextID(context, index+1) == target {
			deleted = true
			continue
		}
		updated = append(updated, context)
	}
	if !deleted {
		return false, nil
	}
	if err := s.localWriteMemoryContexts(updated); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Server) localMemoryQueryResults(query string, limit int) ([]map[string]any, error) {
	memories, err := s.localMemoryExportRecords("default")
	if err != nil {
		return nil, err
	}
	tokens := localUniqueStrings(nil, strings.Fields(strings.ToLower(strings.TrimSpace(query)))...)
	if len(tokens) == 0 && strings.TrimSpace(query) != "" {
		tokens = []string{strings.ToLower(strings.TrimSpace(query))}
	}
	type scoredRecord struct {
		record map[string]any
		score  int
		time   time.Time
	}
	scored := make([]scoredRecord, 0, len(memories))
	for index, memory := range memories {
		record := cloneMap(memory)
		id := localMemoryContextID(record, index+1)
		score := localScoreText(localMemorySearchDocument(record), tokens)
		if strings.TrimSpace(query) != "" && score <= 0 {
			continue
		}
		createdAt, _ := time.Parse(time.RFC3339, localMemoryContextCreatedAt(record))
		metadata, _ := record["metadata"].(map[string]any)
		resultMetadata := cloneMap(metadata)
		resultMetadata["title"] = stringValue(record["title"])
		resultMetadata["source"] = stringValue(record["source"])
		resultMetadata["url"] = stringValue(record["url"])
		resultMetadata["createdAt"] = localMemoryContextCreatedAt(record)
		resultMetadata["userId"] = stringValue(record["userId"])
		resultMetadata["agentId"] = stringValue(record["agentId"])
		scored = append(scored, scoredRecord{
			record: map[string]any{
				"id":       id,
				"content":  stringValue(record["content"]),
				"metadata": resultMetadata,
				"score":    score,
			},
			score: score,
			time:  createdAt,
		})
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].time.After(scored[j].time)
		}
		return scored[i].score > scored[j].score
	})
	if limit <= 0 || limit > len(scored) {
		limit = len(scored)
	}
	results := make([]map[string]any, 0, limit)
	for _, item := range scored[:limit] {
		results = append(results, item.record)
	}
	return results, nil
}
