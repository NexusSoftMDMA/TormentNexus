package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const localAgentMemorySessionTTLMilliseconds = 30 * 60 * 1000

type localAgentMemorySearchOptions struct {
	Limit     int
	Type      string
	Namespace string
	Tags      []string
}

type scoredLocalAgentMemory struct {
	record localAgentMemoryRecord
	score  int
}

func (s *Server) localAgentMemoryFilePath() string {
	return filepath.Join(s.cfg.WorkspaceRoot, ".tormentnexus", "agent_memory", "memories.json")
}

func (s *Server) localWriteAgentMemories(records []localAgentMemoryRecord) error {
	path := s.localAgentMemoryFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	snapshot := map[string]any{
		"version":  1,
		"savedAt":  time.Now().UTC().Format(time.RFC3339Nano),
		"memories": records,
	}
	raw, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func localNewAgentMemoryID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("mem-%d", time.Now().UTC().UnixNano())
}

func localNormalizeAgentMemoryType(value string, defaultValue string) string {
	switch strings.TrimSpace(value) {
	case "session", "working", "long_term":
		return strings.TrimSpace(value)
	default:
		return defaultValue
	}
}

func localNormalizeAgentMemoryNamespace(value string, defaultValue string) string {
	switch strings.TrimSpace(value) {
	case "user", "agent", "project":
		return strings.TrimSpace(value)
	default:
		return defaultValue
	}
}

func localAgentMemoryMetadata(value any) map[string]any {
	metadata, _ := value.(map[string]any)
	if metadata == nil {
		return map[string]any{}
	}
	return cloneMap(metadata)
}

func localAgentMemoryHash(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		raw = []byte(fmt.Sprintf("%v", value))
	}
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}

func localAgentMemoryTags(metadata map[string]any) []string {
	return localUniqueStrings(nil, stringArray(metadata["tags"])...)
}

func localAgentMemoryHasAnyTag(record localAgentMemoryRecord, tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	recordTags := map[string]struct{}{}
	for _, tag := range localAgentMemoryTags(record.Metadata) {
		recordTags[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
	}
	for _, tag := range tags {
		if _, ok := recordTags[strings.ToLower(strings.TrimSpace(tag))]; ok {
			return true
		}
	}
	return false
}

func localAgentMemorySearchDocument(record localAgentMemoryRecord) string {
	parts := []string{record.Content, record.Type, record.Namespace}
	if len(record.Metadata) > 0 {
		if raw, err := json.Marshal(record.Metadata); err == nil {
			parts = append(parts, string(raw))
		}
	}
	return strings.Join(localUniqueStrings(nil, parts...), " ")
}

func localAgentMemoryTokens(query string) []string {
	fields := strings.Fields(strings.ToLower(query))
	if len(fields) == 0 && strings.TrimSpace(query) != "" {
		return []string{strings.ToLower(strings.TrimSpace(query))}
	}
	return localUniqueStrings(nil, fields...)
}

func (s *Server) localFilteredAgentMemoryRecords(memoryType, namespace string, tags []string) ([]localAgentMemoryRecord, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, err
	}

	filtered := make([]localAgentMemoryRecord, 0, len(records))
	for _, record := range records {
		if memoryType != "" && record.Type != memoryType {
			continue
		}
		if namespace != "" && record.Namespace != namespace {
			continue
		}
		if !localAgentMemoryHasAnyTag(record, tags) {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered, nil
}

func (s *Server) localRecentAgentMemoryRecords(limit int, options localAgentMemorySearchOptions) ([]localAgentMemoryRecord, error) {
	records, err := s.localFilteredAgentMemoryRecords(options.Type, options.Namespace, options.Tags)
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		return localAgentMemorySortTime(records[i]).After(localAgentMemorySortTime(records[j]))
	})
	if limit <= 0 || limit > len(records) {
		limit = len(records)
	}
	return append([]localAgentMemoryRecord(nil), records[:limit]...), nil
}

func (s *Server) localSearchAgentMemoryRecords(query string, options localAgentMemorySearchOptions) ([]localAgentMemoryRecord, error) {
	records, err := s.localFilteredAgentMemoryRecords(options.Type, options.Namespace, options.Tags)
	if err != nil {
		return nil, err
	}

	tokens := localAgentMemoryTokens(query)
	scored := make([]scoredLocalAgentMemory, 0, len(records))
	for _, record := range records {
		score := localScoreText(localAgentMemorySearchDocument(record), tokens)
		if strings.TrimSpace(query) != "" && score <= 0 {
			continue
		}
		scored = append(scored, scoredLocalAgentMemory{record: record, score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return localAgentMemorySortTime(scored[i].record).After(localAgentMemorySortTime(scored[j].record))
		}
		return scored[i].score > scored[j].score
	})

	limit := options.Limit
	if limit <= 0 || limit > len(scored) {
		limit = len(scored)
	}
	results := make([]localAgentMemoryRecord, 0, limit)
	for _, item := range scored[:limit] {
		results = append(results, item.record)
	}
	return results, nil
}

func localAgentMemoryMaps(records []localAgentMemoryRecord) []map[string]any {
	mapped := make([]map[string]any, 0, len(records))
	for _, record := range records {
		mapped = append(mapped, localAgentMemoryMap(record))
	}
	return mapped
}

func (s *Server) localAddAgentMemoryEntry(content, memoryType, namespace string, metadata map[string]any) (localAgentMemoryRecord, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return localAgentMemoryRecord{}, fmt.Errorf("missing memory content")
	}

	records, err := s.localAgentMemories()
	if err != nil {
		return localAgentMemoryRecord{}, err
	}

	now := time.Now().UTC()
	record := localAgentMemoryRecord{
		ID:          localNewAgentMemoryID(),
		Content:     content,
		Type:        localNormalizeAgentMemoryType(memoryType, "working"),
		Namespace:   localNormalizeAgentMemoryNamespace(namespace, "project"),
		Metadata:    cloneMap(metadata),
		CreatedAt:   now.Format(time.RFC3339Nano),
		AccessedAt:  now.Format(time.RFC3339Nano),
		AccessCount: 0,
	}
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	if record.Type == "session" {
		ttl := float64(localAgentMemorySessionTTLMilliseconds)
		record.TTL = &ttl
	}

	records = append(records, record)
	if err := s.localWriteAgentMemories(records); err != nil {
		return localAgentMemoryRecord{}, err
	}
	return record, nil
}

func (s *Server) localDeleteAgentMemory(id string) (bool, error) {
	targetID := strings.TrimSpace(id)
	if targetID == "" {
		return false, fmt.Errorf("missing memory id")
	}

	records, err := s.localAgentMemories()
	if err != nil {
		return false, err
	}

	updated := make([]localAgentMemoryRecord, 0, len(records))
	deleted := false
	for _, record := range records {
		if record.ID == targetID {
			deleted = true
			continue
		}
		updated = append(updated, record)
	}
	if !deleted {
		return false, nil
	}
	if err := s.localWriteAgentMemories(updated); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Server) localClearSessionAgentMemories() (int, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return 0, err
	}

	updated := make([]localAgentMemoryRecord, 0, len(records))
	cleared := 0
	for _, record := range records {
		if record.Type == "session" {
			cleared++
			continue
		}
		updated = append(updated, record)
	}
	if err := s.localWriteAgentMemories(updated); err != nil {
		return 0, err
	}
	return cleared, nil
}

func (s *Server) localAgentMemoryExport() (string, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return "", err
	}
	raw, err := json.MarshalIndent(map[string]any{
		"exportedAt": time.Now().UTC().Format(time.RFC3339Nano),
		"memories":   records,
	}, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (s *Server) localAgentMemoryStatsCompact() (map[string]any, error) {
	stats, err := s.localAgentMemoryStats()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"session":  stats["session"],
		"working":  stats["working"],
		"longTerm": stats["long_term"],
		"total":    stats["totalCount"],
	}, nil
}

func (s *Server) localAddFactMemory(payload map[string]any) (map[string]any, error) {
	metadata := localAgentMemoryMetadata(payload["metadata"])
	tags := localUniqueStrings(localAgentMemoryTags(metadata), stringArray(payload["tags"])...)
	metadata["source"] = firstNonEmptyString(stringValue(metadata["source"]), "memory_fact")
	metadata["tags"] = tags
	metadata["memoryKind"] = "fact"

	record, err := s.localAddAgentMemoryEntry(
		stringValue(payload["content"]),
		localNormalizeAgentMemoryType(stringValue(payload["type"]), "working"),
		localNormalizeAgentMemoryNamespace(stringValue(payload["namespace"]), "project"),
		metadata,
	)
	if err != nil {
		return nil, err
	}
	return localAgentMemoryMap(record), nil
}

func localNormalizeObservationType(value string) string {
	switch strings.TrimSpace(value) {
	case "discovery", "decision", "progress", "warning", "fix":
		return strings.TrimSpace(value)
	case "fact":
		return "discovery"
	default:
		return "discovery"
	}
}

func (s *Server) localRecordObservationMemory(payload map[string]any) (map[string]any, error) {
	content := strings.TrimSpace(stringValue(payload["content"]))
	if content == "" {
		content = strings.TrimSpace(stringValue(payload["narrative"]))
	}
	if content == "" {
		return nil, fmt.Errorf("missing observation content")
	}

	metadata := localAgentMemoryMetadata(payload["metadata"])
	observationType := localNormalizeObservationType(stringValue(payload["type"]))
	toolName := strings.TrimSpace(stringValue(payload["toolName"]))
	title := localTrimLine(stringValue(firstNonEmptyString(payload["title"], payload["narrative"], content)), 160)
	narrative := localTrimLine(stringValue(firstNonEmptyString(payload["narrative"], content, title)), 600)
	concepts := localUniqueStrings(nil, stringArray(payload["concepts"])...)
	filesRead := localUniqueStrings(nil, stringArray(payload["filesRead"])...)
	filesModified := localUniqueStrings(nil, stringArray(payload["filesModified"])...)
	recordedAt := float64(time.Now().UTC().UnixMilli())
	contentHash := localAgentMemoryHash(map[string]any{
		"content":       content,
		"title":         title,
		"narrative":     narrative,
		"toolName":      toolName,
		"type":          observationType,
		"concepts":      concepts,
		"filesRead":     filesRead,
		"filesModified": filesModified,
		"rawInput":      stringValue(payload["rawInput"]),
		"rawOutput":     stringValue(payload["rawOutput"]),
	})

	metadata["source"] = firstNonEmptyString(stringValue(metadata["source"]), firstNonEmptyString(toolName, "observation"))
	metadata["tags"] = localUniqueStrings(localAgentMemoryTags(metadata), append([]string{observationType, toolName}, concepts...)...)
	metadata["structuredObservation"] = map[string]any{
		"type":          observationType,
		"title":         title,
		"subtitle":      nullableString(payload["subtitle"]),
		"narrative":     narrative,
		"facts":         localUniqueStrings(nil, stringArray(payload["facts"])...),
		"concepts":      concepts,
		"filesRead":     filesRead,
		"filesModified": filesModified,
		"toolName":      nullableString(toolName),
		"contentHash":   contentHash,
		"recordedAt":    recordedAt,
	}
	metadata["observationType"] = observationType
	metadata["observationHash"] = contentHash
	metadata["toolName"] = nullableString(toolName)
	metadata["filesRead"] = filesRead
	metadata["filesModified"] = filesModified
	if rawInput := strings.TrimSpace(stringValue(payload["rawInput"])); rawInput != "" {
		metadata["rawInput"] = rawInput
	}
	if rawOutput := strings.TrimSpace(stringValue(payload["rawOutput"])); rawOutput != "" {
		metadata["rawOutput"] = rawOutput
	}

	record, err := s.localAddAgentMemoryEntry(content, "working", localNormalizeAgentMemoryNamespace(stringValue(payload["namespace"]), "project"), metadata)
	if err != nil {
		return nil, err
	}
	return localAgentMemoryMap(record), nil
}

func (s *Server) localCaptureUserPromptMemory(payload map[string]any) (map[string]any, error) {
	content := localTrimLine(stringValue(payload["content"]), 400)
	if content == "" {
		return nil, fmt.Errorf("missing user prompt content")
	}
	role := strings.TrimSpace(stringValue(payload["role"]))
	if role == "" {
		role = "prompt"
	}
	metadata := localAgentMemoryMetadata(payload["metadata"])
	recordedAt := float64(time.Now().UTC().UnixMilli())
	contentHash := localAgentMemoryHash(map[string]any{
		"content":       content,
		"role":          role,
		"sessionId":     nullableString(payload["sessionId"]),
		"activeGoal":    nullableString(payload["activeGoal"]),
		"lastObjective": nullableString(payload["lastObjective"]),
	})
	recentPrompts, err := s.localRecentUserPrompts(100, "")
	if err != nil {
		return nil, err
	}
	promptNumber := len(recentPrompts) + 1

	metadata["source"] = firstNonEmptyString(stringValue(metadata["source"]), "user_prompt")
	metadata["sessionId"] = nullableString(payload["sessionId"])
	metadata["tags"] = localUniqueStrings(localAgentMemoryTags(metadata), "user-prompt", role)
	metadata["structuredUserPrompt"] = map[string]any{
		"role":          role,
		"content":       content,
		"sessionId":     nullableString(payload["sessionId"]),
		"activeGoal":    nullableString(payload["activeGoal"]),
		"lastObjective": nullableString(payload["lastObjective"]),
		"promptNumber":  promptNumber,
		"recordedAt":    recordedAt,
		"contentHash":   contentHash,
	}
	metadata["memoryKind"] = "user_prompt"
	metadata["promptRole"] = role

	record, err := s.localAddAgentMemoryEntry(content, "long_term", "project", metadata)
	if err != nil {
		return nil, err
	}
	return localAgentMemoryMap(record), nil
}

func (s *Server) localCaptureSessionSummaryMemory(payload map[string]any) (map[string]any, error) {
	sessionID := strings.TrimSpace(stringValue(payload["sessionId"]))
	status := strings.TrimSpace(stringValue(payload["status"]))
	if sessionID == "" || status == "" {
		return nil, fmt.Errorf("missing sessionId or status")
	}

	metadata := localAgentMemoryMetadata(payload["metadata"])
	name := strings.TrimSpace(stringValue(payload["name"]))
	cliType := strings.TrimSpace(stringValue(payload["cliType"]))
	activeGoal := strings.TrimSpace(stringValue(payload["activeGoal"]))
	lastObjective := strings.TrimSpace(stringValue(payload["lastObjective"]))
	logTail := localUniqueStrings(nil, stringArray(payload["logTail"])...)
	recordedAt := float64(time.Now().UTC().UnixMilli())
	stoppedAt := localNumericValue(payload["stoppedAt"])
	if stoppedAt <= 0 {
		stoppedAt = recordedAt
	}
	contentHash := localAgentMemoryHash(map[string]any{
		"sessionId":     sessionID,
		"name":          name,
		"cliType":       cliType,
		"status":        status,
		"activeGoal":    activeGoal,
		"lastObjective": lastObjective,
		"logTail":       logTail,
	})
	content := strings.TrimSpace(stringValue(payload["summary"]))
	if content == "" {
		label := sessionID
		if name != "" {
			label = name
		}
		content = localTrimLine(label+" ended with status "+status+".", 240)
	}

	metadata["source"] = firstNonEmptyString(stringValue(metadata["source"]), "session_summary")
	metadata["section"] = firstNonEmptyString(stringValue(metadata["section"]), "general")
	metadata["tags"] = localUniqueStrings(localAgentMemoryTags(metadata), "session-summary", cliType, status)
	metadata["sessionId"] = sessionID
	metadata["structuredSessionSummary"] = map[string]any{
		"sessionId":     sessionID,
		"name":          nullableString(name),
		"cliType":       nullableString(cliType),
		"status":        status,
		"activeGoal":    nullableString(activeGoal),
		"lastObjective": nullableString(lastObjective),
		"contentHash":   contentHash,
		"recordedAt":    recordedAt,
		"stoppedAt":     stoppedAt,
		"logTail":       logTail,
	}
	metadata["memoryKind"] = "session_summary"

	record, err := s.localAddAgentMemoryEntry(content, "long_term", "project", metadata)
	if err != nil {
		return nil, err
	}
	return localAgentMemoryMap(record), nil
}

func (s *Server) localSearchObservationMemories(query string, limit int, namespace, observationType string) ([]map[string]any, error) {
	records, err := s.localSearchAgentMemoryRecords(query, localAgentMemorySearchOptions{Limit: limit, Type: "working", Namespace: namespace})
	if err != nil {
		return nil, err
	}
	filtered := make([]localAgentMemoryRecord, 0, len(records))
	for _, record := range records {
		observation, ok := localStructuredObservation(record.Metadata)
		if !ok {
			continue
		}
		if observationType != "" && stringValue(observation["type"]) != observationType {
			continue
		}
		filtered = append(filtered, record)
	}
	return localAgentMemoryMaps(filtered), nil
}

func (s *Server) localSearchUserPromptMemories(query string, limit int, role string) ([]map[string]any, error) {
	records, err := s.localSearchAgentMemoryRecords(query, localAgentMemorySearchOptions{Limit: limit, Type: "long_term", Namespace: "project", Tags: []string{"user-prompt"}})
	if err != nil {
		return nil, err
	}
	filtered := make([]localAgentMemoryRecord, 0, len(records))
	for _, record := range records {
		prompt, ok := localStructuredUserPrompt(record.Metadata)
		if !ok {
			continue
		}
		if role != "" && stringValue(prompt["role"]) != role {
			continue
		}
		filtered = append(filtered, record)
	}
	return localAgentMemoryMaps(filtered), nil
}

func (s *Server) localSearchSessionSummaryMemories(query string, limit int) ([]map[string]any, error) {
	records, err := s.localSearchAgentMemoryRecords(query, localAgentMemorySearchOptions{Limit: limit, Type: "long_term", Namespace: "project", Tags: []string{"session-summary"}})
	if err != nil {
		return nil, err
	}
	filtered := make([]localAgentMemoryRecord, 0, len(records))
	for _, record := range records {
		if _, ok := localStructuredSessionSummary(record.Metadata); !ok {
			continue
		}
		filtered = append(filtered, record)
	}
	return localAgentMemoryMaps(filtered), nil
}
