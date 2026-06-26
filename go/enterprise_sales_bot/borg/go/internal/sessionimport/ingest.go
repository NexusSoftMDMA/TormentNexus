package sessionimport

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type IngestSummary struct {
	DiscoveredCount    int      `json:"discoveredCount"`
	ImportedCount      int      `json:"importedCount"`
	SkippedCount       int      `json:"skippedCount"`
	StoredMemoryCount  int      `json:"storedMemoryCount"`
	InstructionDocPath *string  `json:"instructionDocPath"`
	Tools              []string `json:"tools"`
	Errors             []string `json:"errors,omitempty"`
}

func IngestDiscoveredSessions(ctx context.Context, workspaceRoot, homeDir string, maxFiles int, force bool) (*IngestSummary, error) {
	scanner := NewScanner(workspaceRoot, homeDir, maxFiles)
	candidates, err := scanner.Scan()
	if err != nil {
		return nil, err
	}
	store := NewImportedSessionStore(workspaceRoot)
	summary := &IngestSummary{DiscoveredCount: len(candidates), Errors: []string{}}
	toolsSet := map[string]struct{}{}

	for _, candidate := range candidates {
		toolsSet[candidate.SourceTool] = struct{}{}
		validation := ValidateCandidate(candidate)
		if !validation.Valid {
			summary.SkippedCount++
			if len(validation.Errors) > 0 {
				summary.Errors = append(summary.Errors, fmt.Sprintf("skip %s: %s", candidate.SourcePath, strings.Join(validation.Errors, "; ")))
			}
			continue
		}
		recordInputs, err := buildImportedSessionRecordInputs(candidate, validation, ctx)
		if err != nil {
			summary.SkippedCount++
			summary.Errors = append(summary.Errors, fmt.Sprintf("skip %s: %v", candidate.SourcePath, err))
			continue
		}
		if len(recordInputs) == 0 {
			summary.SkippedCount++
			continue
		}
		importedForCandidate := 0
		for _, recordInput := range recordInputs {
			if !force {
				exists, err := store.HasTranscriptHash(ctx, recordInput.TranscriptHash)
				if err != nil {
					summary.Errors = append(summary.Errors, fmt.Sprintf("dedup %s: %v", recordInput.SourcePath, err))
					continue
				}
				if exists {
					summary.SkippedCount++
					continue
				}
			}

			record, err := store.UpsertSession(ctx, recordInput)
			if err != nil {
				summary.Errors = append(summary.Errors, fmt.Sprintf("persist %s: %v", recordInput.SourcePath, err))
				continue
			}
			importedForCandidate++
			summary.ImportedCount++
			summary.StoredMemoryCount += len(record.ParsedMemories)
		}
		if importedForCandidate == 0 {
			summary.SkippedCount++
		}
	}

	doc, err := store.WriteInstructionDoc(ctx, 250)
	if err != nil {
		return nil, err
	}
	if doc != nil {
		summary.InstructionDocPath = &doc.Path
	}

	summary.Tools = make([]string, 0, len(toolsSet))
	for tool := range toolsSet {
		summary.Tools = append(summary.Tools, tool)
	}
	sort.Strings(summary.Tools)
	return summary, nil
}

func buildImportedSessionRecordInputs(candidate Candidate, validation ValidationResult, ctx context.Context) ([]ImportedSessionRecordInput, error) {
	if validation.SourceType == "database-log" || candidate.SessionFormat == "db" {
		return buildImportedSessionRecordInputsFromDatabase(ctx, candidate)
	}
	recordInput, err := buildImportedSessionRecordInput(candidate, validation)
	if err != nil {
		return nil, err
	}
	return []ImportedSessionRecordInput{recordInput}, nil
}

func buildImportedSessionRecordInput(candidate Candidate, validation ValidationResult) (ImportedSessionRecordInput, error) {
	content, err := os.ReadFile(candidate.SourcePath)
	if err != nil {
		return ImportedSessionRecordInput{}, err
	}
	transcript := strings.TrimSpace(parseTranscriptContent(candidate.SourcePath, content))
	if transcript == "" {
		return ImportedSessionRecordInput{}, fmt.Errorf("empty transcript after parsing")
	}
	transcriptHash := hashTranscript(transcript)
	lines := normalizedLines(transcript)
	now := time.Now().UTC().UnixMilli()
	lastModifiedAt := parseCandidateTimestamp(candidate.LastModifiedAt)
	sourcePath := candidate.SourcePath
	title := firstNonEmptyString(filepath.Base(sourcePath), firstString(lines))
	if title == "" {
		title = "Imported Session"
	}
	excerpt := transcript
	if len(excerpt) > 320 {
		excerpt = excerpt[:320]
	}
	workingDirectory := filepath.Dir(sourcePath)
	externalSessionID := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	sessionID := "imported:" + candidate.SourceTool + ":" + transcriptHash[:12]
	memories := heuristicMemoryExtraction(transcript, candidate.SourceTool)
	retentionSummary := buildRetentionSummary(transcript, memories)
	for index := range memories {
		if memories[index].Metadata == nil {
			memories[index].Metadata = map[string]any{}
		}
		memories[index].Metadata["sourceTool"] = candidate.SourceTool
		memories[index].Metadata["path"] = sourcePath
		memories[index].Metadata["sessionId"] = sessionID
	}

	return ImportedSessionRecordInput{
		SourceTool:        candidate.SourceTool,
		SourcePath:        sourcePath,
		ExternalSessionID: stringPtr(externalSessionID),
		Title:             stringPtr(title),
		SessionFormat:     candidate.SessionFormat,
		Transcript:        transcript,
		Excerpt:           stringPtr(excerpt),
		WorkingDirectory:  stringPtr(workingDirectory),
		TranscriptHash:    transcriptHash,
		NormalizedSession: map[string]any{
			"sessionId":         sessionID,
			"title":             title,
			"sourceTool":        candidate.SourceTool,
			"sourcePath":        sourcePath,
			"sessionFormat":     candidate.SessionFormat,
			"externalSessionId": externalSessionID,
			"transcriptHash":    transcriptHash,
			"lineCount":         len(lines),
			"contentLength":     len(transcript),
			"excerpt":           excerpt,
			"importedAt":        now,
			"lastModifiedAt":    nullableTimeValue(lastModifiedAt),
			"detectedModels":    validation.DetectedModels,
		},
		Metadata: map[string]any{
			"sourceTool":       candidate.SourceTool,
			"sourcePath":       sourcePath,
			"sessionFormat":    candidate.SessionFormat,
			"contentLength":    len(transcript),
			"lineCount":        len(lines),
			"sourceType":       validation.SourceType,
			"detectedModels":   validation.DetectedModels,
			"retentionSummary": retentionSummary,
		},
		DiscoveredAt:   now,
		ImportedAt:     now,
		LastModifiedAt: lastModifiedAt,
		ParsedMemories: memories,
	}, nil
}

func parseTranscriptContent(filePath string, content []byte) string {
	extension := strings.ToLower(filepath.Ext(filePath))
	text := string(content)
	switch extension {
	case ".jsonl":
		lines := []string{}
		scanner := bufio.NewScanner(strings.NewReader(text))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var parsed any
			if err := json.Unmarshal([]byte(line), &parsed); err == nil {
				lines = append(lines, extractJSONTextFragments(parsed)...)
			} else {
				lines = append(lines, line)
			}
		}
		return strings.Join(uniqueStrings(lines, 400), "\n")
	case ".json":
		var parsed any
		if err := json.Unmarshal(content, &parsed); err == nil {
			fragments := uniqueStrings(extractJSONTextFragments(parsed), 400)
			if len(fragments) > 0 {
				return strings.Join(fragments, "\n")
			}
		}
	}
	return text
}

func extractJSONTextFragments(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []any:
		result := []string{}
		for _, item := range typed {
			result = append(result, extractJSONTextFragments(item)...)
		}
		return result
	case map[string]any:
		result := []string{}
		keys := []string{"content", "text", "message", "prompt", "response", "request", "input", "output", "result", "body", "summary", "title", "parts", "messages", "conversation", "transcript", "entries", "events", "turns", "items"}
		for _, key := range keys {
			if nested, ok := typed[key]; ok {
				result = append(result, extractJSONTextFragments(nested)...)
			}
		}
		return result
	default:
		return nil
	}
}

func heuristicMemoryExtraction(text string, sourceTool string) []ImportedSessionMemoryInput {
	candidateLines := []string{}
	for _, line := range normalizedLines(text) {
		sanitized := sanitizeSentence(line, 240)
		if len(sanitized) < 24 {
			continue
		}
		if !memoryHintPattern.MatchString(strings.ToLower(sanitized)) {
			continue
		}
		candidateLines = append(candidateLines, sanitized)
	}
	candidateSentences := []string{}
	for _, sentence := range splitSentences(text) {
		sanitized := sanitizeSentence(sentence, 220)
		if len(sanitized) < 30 {
			continue
		}
		if !memoryHintPattern.MatchString(strings.ToLower(sanitized)) {
			continue
		}
		candidateSentences = append(candidateSentences, sanitized)
	}
	facts := uniqueStrings(append(candidateLines, candidateSentences...), 6)
	result := make([]ImportedSessionMemoryInput, 0, len(facts))
	for _, fact := range facts {
		result = append(result, ImportedSessionMemoryInput{
			Kind:    classifyMemoryKind(fact),
			Content: fact,
			Tags:    deriveTags(fact, sourceTool),
			Source:  ImportedSessionMemorySourceHeuristic,
			Metadata: map[string]any{
				"extraction": "heuristic",
			},
		})
	}
	return result
}

var memoryHintPattern = regexp.MustCompile(`use|prefer|should|must|avoid|remember|fixed|fix|discovered|default|path|port|error|warning|supports|requires`)
var instructionPattern = regexp.MustCompile(`always|never|prefer|should|must|avoid|remember to|do not|don't|use\b`)

func classifyMemoryKind(content string) ImportedSessionMemoryKind {
	if instructionPattern.MatchString(strings.ToLower(content)) {
		return ImportedSessionMemoryKindInstruction
	}
	return ImportedSessionMemoryKindMemory
}

func deriveTags(content string, sourceTool string) []string {
	lowered := strings.ToLower(content)
	tags := []string{sourceTool, string(classifyMemoryKind(content))}
	if strings.Contains(lowered, "port") || strings.Contains(lowered, "localhost") || strings.Contains(lowered, "127.0.0.1") || strings.Contains(lowered, "http") || strings.Contains(lowered, "ws:") || strings.Contains(lowered, "wss:") {
		tags = append(tags, "networking")
	}
	if strings.Contains(lowered, "memory") || strings.Contains(lowered, "context") || strings.Contains(lowered, "session") || strings.Contains(lowered, "history") {
		tags = append(tags, "memory")
	}
	if strings.Contains(lowered, "sqlite") || strings.Contains(lowered, "database") || strings.Contains(lowered, " db") {
		tags = append(tags, "database")
	}
	if strings.Contains(lowered, "build") || strings.Contains(lowered, "typecheck") || strings.Contains(lowered, "test") || strings.Contains(lowered, "vitest") || strings.Contains(lowered, "tsc") {
		tags = append(tags, "validation")
	}
	if strings.Contains(lowered, "mcp") || strings.Contains(lowered, "tool") || strings.Contains(lowered, "server") || strings.Contains(lowered, "catalog") {
		tags = append(tags, "mcp")
	}
	if strings.Contains(lowered, "dashboard") || strings.Contains(lowered, "ui") || strings.Contains(lowered, "widget") || strings.Contains(lowered, "page") {
		tags = append(tags, "ui")
	}
	return uniqueStrings(tags, 8)
}

func buildRetentionSummary(transcript string, memories []ImportedSessionMemoryInput) map[string]any {
	durableInstructionCount := 0
	for _, memory := range memories {
		if memory.Kind == ImportedSessionMemoryKindInstruction {
			durableInstructionCount++
		}
	}
	durableMemoryCount := len(memories) - durableInstructionCount
	summary := "Archived full transcript for historical reference only; no durable memories were promoted."
	if len(memories) > 0 {
		summary = fmt.Sprintf("Archived full transcript; promoted %d durable item(s) to fast memory while keeping the remaining context compressed.", len(memories))
	}
	return map[string]any{
		"strategy":                "heuristic",
		"transcriptLength":        len(transcript),
		"analyzedChars":           len(transcript),
		"durableMemoryCount":      durableMemoryCount,
		"durableInstructionCount": durableInstructionCount,
		"archiveDisposition":      "archive_only",
		"summary":                 sanitizeSentence(summary, 240),
		"salientTags":             uniqueStrings(flattenMemoryTags(memories), 10),
		"keepArchivedCategories":  []string{"conversational-context", "implementation-detail", "historical-trace"},
		"discardableCategories":   []string{"greetings", "small-talk", "duplicate-restatements"},
	}
}

func flattenMemoryTags(memories []ImportedSessionMemoryInput) []string {
	result := []string{}
	for _, memory := range memories {
		result = append(result, memory.Tags...)
	}
	return result
}

func normalizedLines(text string) []string {
	parts := strings.Split(text, "\n")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitSentences(text string) []string {
	replacer := strings.NewReplacer("\r", " ", "\n", " ")
	normalized := replacer.Replace(text)
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func sanitizeSentence(value string, maxLength int) string {
	trimmed := strings.Join(strings.Fields(value), " ")
	if len(trimmed) <= maxLength {
		return trimmed
	}
	return trimmed[:maxLength]
}

func uniqueStrings(values []string, maxItems int) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
		if maxItems > 0 && len(result) >= maxItems {
			break
		}
	}
	return result
}

func hashTranscript(transcript string) string {
	hasher := sha256.Sum256([]byte(transcript))
	return hex.EncodeToString(hasher[:])
}

func parseCandidateTimestamp(raw string) *int64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	value := parsed.UTC().UnixMilli()
	return &value
}

func nullableTimeValue(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
