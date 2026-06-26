package httpapi

import (
	"encoding/json"
	"sort"
	"strings"
	"time"
)

var localIntentStopWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "been": {}, "being": {},
	"but": {}, "by": {}, "for": {}, "from": {}, "had": {}, "has": {}, "have": {}, "in": {}, "into": {},
	"is": {}, "it": {}, "its": {}, "of": {}, "on": {}, "or": {}, "that": {}, "the": {}, "their": {},
	"them": {}, "these": {}, "this": {}, "those": {}, "to": {}, "was": {}, "were": {}, "with": {},
}

type localCrossSessionMemoryLink struct {
	Record  localAgentMemoryRecord
	Score   int
	Reasons []string
}

func localFindAgentMemoryRecord(records []localAgentMemoryRecord, id string) (localAgentMemoryRecord, bool) {
	target := strings.TrimSpace(id)
	if target == "" {
		return localAgentMemoryRecord{}, false
	}
	for _, record := range records {
		if record.ID == target {
			return record, true
		}
	}
	return localAgentMemoryRecord{}, false
}

func localAgentMemoryMapWithScore(record localAgentMemoryRecord, score int) map[string]any {
	mapped := localAgentMemoryMap(record)
	mapped["score"] = score
	return mapped
}

func localMemorySessionID(record localAgentMemoryRecord) string {
	if structured, ok := localStructuredSessionSummary(record.Metadata); ok {
		if sessionID := strings.TrimSpace(stringValue(structured["sessionId"])); sessionID != "" {
			return sessionID
		}
	}
	if structured, ok := localStructuredUserPrompt(record.Metadata); ok {
		if sessionID := strings.TrimSpace(stringValue(structured["sessionId"])); sessionID != "" {
			return sessionID
		}
	}
	return strings.TrimSpace(stringValue(record.Metadata["sessionId"]))
}

func localMemoryToolName(record localAgentMemoryRecord) string {
	if structured, ok := localStructuredObservation(record.Metadata); ok {
		return strings.TrimSpace(stringValue(structured["toolName"]))
	}
	return ""
}

func localMemoryConcepts(record localAgentMemoryRecord) []string {
	if structured, ok := localStructuredObservation(record.Metadata); ok {
		return localUniqueStrings(nil, stringArray(structured["concepts"])...)
	}
	return []string{}
}

func localMemoryFiles(record localAgentMemoryRecord) []string {
	if structured, ok := localStructuredObservation(record.Metadata); ok {
		values := append(stringArray(structured["filesRead"]), stringArray(structured["filesModified"])...)
		normalized := make([]string, 0, len(values))
		for _, value := range values {
			value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
			if value != "" {
				normalized = append(normalized, value)
			}
		}
		return localUniqueStrings(nil, normalized...)
	}
	return []string{}
}

func localMemoryGoalSignals(record localAgentMemoryRecord) []string {
	values := make([]string, 0, 3)
	if structured, ok := localStructuredSessionSummary(record.Metadata); ok {
		values = append(values, stringValue(structured["activeGoal"]))
	}
	if structured, ok := localStructuredUserPrompt(record.Metadata); ok {
		values = append(values, stringValue(structured["activeGoal"]))
		if strings.TrimSpace(stringValue(structured["role"])) == "goal" {
			values = append(values, record.Content)
		}
	}
	return localNonEmptyLowercaseSignals(values)
}

func localMemoryObjectiveSignals(record localAgentMemoryRecord) []string {
	values := make([]string, 0, 3)
	if structured, ok := localStructuredSessionSummary(record.Metadata); ok {
		values = append(values, stringValue(structured["lastObjective"]))
	}
	if structured, ok := localStructuredUserPrompt(record.Metadata); ok {
		values = append(values, stringValue(structured["lastObjective"]))
		if strings.TrimSpace(stringValue(structured["role"])) == "objective" {
			values = append(values, record.Content)
		}
	}
	return localNonEmptyLowercaseSignals(values)
}

func localNonEmptyLowercaseSignals(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		candidate := strings.TrimSpace(strings.ToLower(value))
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		normalized = append(normalized, candidate)
	}
	return normalized
}

func localNormalizePivotValue(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"))
}

func localScoreDirectPivotMatch(record localAgentMemoryRecord, pivot, value string) int {
	switch pivot {
	case "session":
		if sessionID := strings.ToLower(localMemorySessionID(record)); sessionID != "" && sessionID == value {
			return 10
		}
	case "tool":
		if toolName := strings.ToLower(localMemoryToolName(record)); toolName != "" && toolName == value {
			return 10
		}
	case "concept":
		for _, concept := range localMemoryConcepts(record) {
			if strings.ToLower(concept) == value {
				return 10
			}
		}
	case "goal":
		for _, goal := range localMemoryGoalSignals(record) {
			if goal == value {
				return 10
			}
		}
	case "objective":
		for _, objective := range localMemoryObjectiveSignals(record) {
			if objective == value {
				return 10
			}
		}
	case "file":
		for _, file := range localMemoryFiles(record) {
			if strings.ToLower(file) == value {
				return 10
			}
		}
	}
	return 0
}

func localInferPivotFromAnchor(record localAgentMemoryRecord) (string, string, bool) {
	if sessionID := strings.TrimSpace(localMemorySessionID(record)); sessionID != "" {
		return "session", sessionID, true
	}
	if toolName := strings.TrimSpace(localMemoryToolName(record)); toolName != "" {
		return "tool", toolName, true
	}
	concepts := localMemoryConcepts(record)
	if len(concepts) > 0 {
		return "concept", concepts[0], true
	}
	files := localMemoryFiles(record)
	if len(files) > 0 {
		return "file", files[0], true
	}
	goals := localMemoryGoalSignals(record)
	if len(goals) > 0 {
		return "goal", goals[0], true
	}
	objectives := localMemoryObjectiveSignals(record)
	if len(objectives) > 0 {
		return "objective", objectives[0], true
	}
	return "", "", false
}

func (s *Server) localSearchMemoryPivotPayload(payload map[string]any) ([]map[string]any, string, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, "", err
	}

	pivot := strings.TrimSpace(stringValue(payload["pivot"]))
	value := strings.TrimSpace(stringValue(payload["value"]))
	limit := int(localNumericValue(payload["limit"]))
	if limit <= 0 {
		limit = 20
	}
	reason := "upstream unavailable; using local persisted pivot search"

	if pivot == "" || value == "" {
		anchorID := strings.TrimSpace(stringValue(payload["pivotMemoryId"]))
		anchor, ok := localFindAgentMemoryRecord(records, anchorID)
		if !ok {
			return []map[string]any{}, "upstream unavailable; local memory pivot anchor was not found", nil
		}
		inferredPivot, inferredValue, ok := localInferPivotFromAnchor(anchor)
		if !ok {
			return []map[string]any{}, "upstream unavailable; local memory pivot fallback could not infer a pivot from the anchor memory", nil
		}
		pivot = inferredPivot
		value = inferredValue
		reason = "upstream unavailable; using local inferred pivot search from anchor memory"
	}

	normalizedValue := localNormalizePivotValue(value)
	if pivot == "" || normalizedValue == "" {
		return []map[string]any{}, "upstream unavailable; local memory pivot search received no usable pivot input", nil
	}

	directMatches := map[string]scoredLocalAgentMemory{}
	relatedSessionIDs := map[string]struct{}{}
	for _, record := range records {
		score := localScoreDirectPivotMatch(record, pivot, normalizedValue)
		if score <= 0 {
			continue
		}
		directMatches[record.ID] = scoredLocalAgentMemory{record: record, score: score}
		if sessionID := strings.ToLower(strings.TrimSpace(localMemorySessionID(record))); sessionID != "" {
			relatedSessionIDs[sessionID] = struct{}{}
		}
	}

	if pivot != "session" && len(relatedSessionIDs) > 0 {
		for _, record := range records {
			if _, exists := directMatches[record.ID]; exists {
				continue
			}
			sessionID := strings.ToLower(strings.TrimSpace(localMemorySessionID(record)))
			if sessionID == "" {
				continue
			}
			if _, ok := relatedSessionIDs[sessionID]; !ok {
				continue
			}
			directMatches[record.ID] = scoredLocalAgentMemory{record: record, score: 4}
		}
	}

	matches := make([]scoredLocalAgentMemory, 0, len(directMatches))
	for _, match := range directMatches {
		matches = append(matches, match)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return localAgentMemorySortTime(matches[i].record).After(localAgentMemorySortTime(matches[j].record))
		}
		return matches[i].score > matches[j].score
	})
	if limit > len(matches) {
		limit = len(matches)
	}
	results := make([]map[string]any, 0, limit)
	for _, match := range matches[:limit] {
		results = append(results, localAgentMemoryMapWithScore(match.record, match.score))
	}
	return results, reason, nil
}

func (s *Server) localTimelineWindowPayload(payload map[string]any) ([]map[string]any, string, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, "", err
	}

	sessionID := strings.TrimSpace(stringValue(payload["sessionId"]))
	anchorTimestamp := localNumericValue(payload["anchorTimestamp"])
	reason := "upstream unavailable; using local persisted memory timeline window"
	if sessionID == "" || anchorTimestamp <= 0 {
		anchorID := strings.TrimSpace(stringValue(payload["centerMemoryId"]))
		anchor, ok := localFindAgentMemoryRecord(records, anchorID)
		if !ok {
			return []map[string]any{}, "upstream unavailable; local timeline anchor was not found", nil
		}
		sessionID = localMemorySessionID(anchor)
		anchorTimestamp = float64(localTimeToMillis(localAgentMemorySortTime(anchor)))
		if sessionID == "" || anchorTimestamp <= 0 {
			return []map[string]any{}, "upstream unavailable; local timeline fallback could not infer session context from the anchor memory", nil
		}
		reason = "upstream unavailable; using local timeline window inferred from anchor memory"
	}

	before := int(localNumericValue(payload["before"]))
	after := int(localNumericValue(payload["after"]))
	if before < 0 {
		before = 3
	}
	if after < 0 {
		after = 3
	}
	normalizedSessionID := strings.ToLower(strings.TrimSpace(sessionID))
	sessionRecords := make([]localAgentMemoryRecord, 0)
	for _, record := range records {
		if strings.ToLower(strings.TrimSpace(localMemorySessionID(record))) == normalizedSessionID {
			sessionRecords = append(sessionRecords, record)
		}
	}
	if len(sessionRecords) == 0 {
		return []map[string]any{}, "upstream unavailable; local memory timeline has no records for the requested session", nil
	}
	sort.Slice(sessionRecords, func(i, j int) bool {
		return localAgentMemorySortTime(sessionRecords[i]).Before(localAgentMemorySortTime(sessionRecords[j]))
	})
	anchorIndex := -1
	for index, record := range sessionRecords {
		if float64(localTimeToMillis(localAgentMemorySortTime(record))) >= anchorTimestamp {
			anchorIndex = index
			break
		}
	}
	if anchorIndex == -1 {
		anchorIndex = len(sessionRecords) - 1
	} else if anchorIndex > 0 {
		candidate := sessionRecords[anchorIndex]
		previous := sessionRecords[anchorIndex-1]
		candidateDistance := absInt64(int64(localTimeToMillis(localAgentMemorySortTime(candidate))) - int64(anchorTimestamp))
		previousDistance := absInt64(int64(localTimeToMillis(localAgentMemorySortTime(previous))) - int64(anchorTimestamp))
		if previousDistance <= candidateDistance {
			anchorIndex--
		}
	}
	start := maxInt(0, anchorIndex-before)
	end := minInt(len(sessionRecords), anchorIndex+after+1)
	return localAgentMemoryMaps(sessionRecords[start:end]), reason, nil
}

func localUniqueOverlap(left, right []string) []string {
	rightSet := map[string]struct{}{}
	for _, value := range right {
		rightSet[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
	}
	overlap := make([]string, 0)
	seen := map[string]struct{}{}
	for _, value := range left {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := rightSet[normalized]; !ok {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		overlap = append(overlap, value)
	}
	return overlap
}

func localNormalizeIntentToken(token string) string {
	trimmed := strings.ToLower(strings.TrimSpace(token))
	if trimmed == "" {
		return ""
	}
	if len(trimmed) > 6 && strings.HasSuffix(trimmed, "ing") {
		return trimmed[:len(trimmed)-3]
	}
	if len(trimmed) > 5 && strings.HasSuffix(trimmed, "ed") {
		return trimmed[:len(trimmed)-2]
	}
	if len(trimmed) > 5 && strings.HasSuffix(trimmed, "es") {
		return trimmed[:len(trimmed)-2]
	}
	if len(trimmed) > 4 && strings.HasSuffix(trimmed, "s") {
		return trimmed[:len(trimmed)-1]
	}
	return trimmed
}

func localIntentTokens(value string) []string {
	rawTokens := strings.FieldsFunc(value, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
	})
	seen := map[string]struct{}{}
	tokens := make([]string, 0, len(rawTokens))
	for _, raw := range rawTokens {
		normalized := localNormalizeIntentToken(raw)
		if len(normalized) < 4 {
			continue
		}
		if _, stop := localIntentStopWords[normalized]; stop {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		tokens = append(tokens, normalized)
	}
	return tokens
}

func localSessionGoalSignals(records []localAgentMemoryRecord, sessionID string) []string {
	trimmed := strings.TrimSpace(sessionID)
	if trimmed == "" {
		return []string{}
	}
	seen := map[string]struct{}{}
	signals := make([]string, 0)
	for _, record := range records {
		if strings.TrimSpace(localMemorySessionID(record)) != trimmed {
			continue
		}
		for _, signal := range append(localMemoryGoalSignals(record), localMemoryObjectiveSignals(record)...) {
			if _, ok := seen[signal]; ok {
				continue
			}
			seen[signal] = struct{}{}
			signals = append(signals, signal)
		}
	}
	return signals
}

func localIntentThemeReasons(leftSignals, rightSignals []string) []string {
	reasons := make([]string, 0)
	seen := map[string]struct{}{}
	for _, left := range leftSignals {
		leftTokens := localIntentTokens(left)
		if len(leftTokens) < 2 {
			continue
		}
		for _, right := range rightSignals {
			if left == right {
				continue
			}
			rightTokens := localIntentTokens(right)
			if len(rightTokens) < 2 {
				continue
			}
			shared := localUniqueOverlap(leftTokens, rightTokens)
			overlapRatio := 0.0
			if minLen := minInt(len(leftTokens), len(rightTokens)); minLen > 0 {
				overlapRatio = float64(len(shared)) / float64(minLen)
			}
			containsOther := strings.Contains(left, right) || strings.Contains(right, left)
			if len(shared) < 2 && !containsOther {
				continue
			}
			if !containsOther && overlapRatio < 0.5 {
				continue
			}
			key := left + "::" + right
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			reason := "similar goal/objective theme"
			if len(shared) > 0 {
				reason += ": " + strings.Join(shared[:minInt(2, len(shared))], ", ")
			}
			reasons = append(reasons, reason)
		}
	}
	return reasons
}

func localCrossSessionLinkPayload(records []localAgentMemoryRecord, memoryID string, limit int) []localCrossSessionMemoryLink {
	anchor, ok := localFindAgentMemoryRecord(records, memoryID)
	if !ok {
		return []localCrossSessionMemoryLink{}
	}
	if limit <= 0 {
		limit = 5
	}
	anchorSessionID := localMemorySessionID(anchor)
	anchorToolName := localMemoryToolName(anchor)
	anchorSource := strings.TrimSpace(stringValue(anchor.Metadata["source"]))
	anchorConcepts := localMemoryConcepts(anchor)
	anchorFiles := localMemoryFiles(anchor)
	anchorGoalSignals := localSessionGoalSignals(records, anchorSessionID)

	related := make([]localCrossSessionMemoryLink, 0)
	for _, candidate := range records {
		if candidate.ID == anchor.ID {
			continue
		}
		candidateSessionID := localMemorySessionID(candidate)
		if candidateSessionID == "" || (anchorSessionID != "" && candidateSessionID == anchorSessionID) {
			continue
		}
		score := 0
		reasons := make([]string, 0)

		sharedConcepts := localUniqueOverlap(anchorConcepts, localMemoryConcepts(candidate))
		if len(sharedConcepts) > 0 {
			score += minInt(len(sharedConcepts), 2) * 3
			reasons = append(reasons, "shared concepts: "+strings.Join(sharedConcepts[:minInt(2, len(sharedConcepts))], ", "))
		}

		sharedFiles := localUniqueOverlap(anchorFiles, localMemoryFiles(candidate))
		if len(sharedFiles) > 0 {
			score += minInt(len(sharedFiles), 2) * 3
			reasons = append(reasons, "shared file: "+sharedFiles[0])
		}

		candidateGoalSignals := localSessionGoalSignals(records, candidateSessionID)
		sharedGoals := localUniqueOverlap(anchorGoalSignals, candidateGoalSignals)
		if len(sharedGoals) > 0 {
			score += minInt(len(sharedGoals), 2) * 4
			reasons = append(reasons, "shared goal/objective: "+sharedGoals[0])
		} else {
			themeReasons := localIntentThemeReasons(anchorGoalSignals, candidateGoalSignals)
			if len(themeReasons) > 0 {
				score += minInt(len(themeReasons), 2) * 3
				reasons = append(reasons, themeReasons[0])
			}
		}

		candidateToolName := localMemoryToolName(candidate)
		if anchorToolName != "" && candidateToolName != "" && anchorToolName == candidateToolName {
			score += 2
			reasons = append(reasons, "same tool ("+anchorToolName+")")
		}
		candidateSource := strings.TrimSpace(stringValue(candidate.Metadata["source"]))
		if anchorSource != "" && candidateSource != "" && anchorSource == candidateSource {
			score += 1
			reasons = append(reasons, "same source ("+anchorSource+")")
		}
		if score > 0 {
			reasons = append(reasons, "other session ("+candidateSessionID+")")
			related = append(related, localCrossSessionMemoryLink{Record: candidate, Score: score, Reasons: reasons})
		}
	}
	sort.Slice(related, func(i, j int) bool {
		if related[i].Score == related[j].Score {
			return localAgentMemorySortTime(related[i].Record).After(localAgentMemorySortTime(related[j].Record))
		}
		return related[i].Score > related[j].Score
	})
	if limit > len(related) {
		limit = len(related)
	}
	return related[:limit]
}

func (s *Server) localCrossSessionMemoryLinksPayload(payload map[string]any) ([]map[string]any, string, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, "", err
	}
	memoryID := strings.TrimSpace(stringValue(payload["memoryId"]))
	limit := int(localNumericValue(payload["limit"]))
	links := localCrossSessionLinkPayload(records, memoryID, limit)
	results := make([]map[string]any, 0, len(links))
	for _, link := range links {
		results = append(results, map[string]any{
			"memory":  localAgentMemoryMap(link.Record),
			"score":   link.Score,
			"reasons": append([]string(nil), link.Reasons...),
		})
	}
	reason := "upstream unavailable; using local persisted cross-session memory links"
	if strings.TrimSpace(memoryID) == "" {
		reason = "upstream unavailable; local cross-session link search received no memory id"
	}
	if strings.TrimSpace(memoryID) != "" && len(results) == 0 {
		reason = "upstream unavailable; local cross-session link search found no related persisted memories"
	}
	return results, reason, nil
}

func (s *Server) localAgentMemoryHandoffArtifact(payload map[string]any) (string, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return "", err
	}
	stats, err := s.localAgentMemoryStats()
	if err != nil {
		return "", err
	}
	sessionRecords := make([]localAgentMemoryRecord, 0)
	for _, record := range records {
		if record.Type == "session" {
			sessionRecords = append(sessionRecords, record)
		}
	}
	sort.Slice(sessionRecords, func(i, j int) bool {
		return localAgentMemorySortTime(sessionRecords[i]).After(localAgentMemorySortTime(sessionRecords[j]))
	})
	if len(sessionRecords) > 20 {
		sessionRecords = sessionRecords[:20]
	}
	recentContext := make([]map[string]any, 0, len(sessionRecords))
	for _, record := range sessionRecords {
		recentContext = append(recentContext, map[string]any{
			"content":  record.Content,
			"metadata": cloneMap(record.Metadata),
		})
	}
	artifact := map[string]any{
		"version":       "0.99.1",
		"timestamp":     time.Now().UTC().UnixMilli(),
		"sessionId":     firstNonEmptyString(stringValue(payload["sessionId"]), "current"),
		"stats":         stats,
		"recentContext": recentContext,
		"notes":         stringValue(payload["notes"]),
	}
	raw, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (s *Server) localAgentMemoryPickupArtifact(artifact string) (map[string]any, error) {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(artifact), &parsed); err != nil {
		return map[string]any{"success": false, "count": 0}, nil
	}
	recentContext, _ := parsed["recentContext"].([]any)
	count := 0
	for _, item := range recentContext {
		entry, _ := item.(map[string]any)
		content := stringValue(entry["content"])
		metadata := localAgentMemoryMetadata(entry["metadata"])
		if strings.TrimSpace(content) == "" {
			continue
		}
		if _, err := s.localAddAgentMemoryEntry(content, "session", "project", metadata); err != nil {
			return nil, err
		}
		count++
	}
	return map[string]any{"success": true, "count": count}, nil
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
