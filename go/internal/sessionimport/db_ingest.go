package sessionimport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func buildImportedSessionRecordInputsFromDatabase(ctx context.Context, candidate Candidate) ([]ImportedSessionRecordInput, error) {
	db, err := sql.Open("sqlite", candidate.SourcePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	switch candidate.SourceTool {
	case "llm-cli":
		return buildLLMDatabaseRecordInputs(ctx, db, candidate)
	case "tormentnexus-mcp":
		return buildTormentNexusDatabaseRecordInputs(ctx, db, candidate)
	default:
		return nil, fmt.Errorf("database-log ingestion is not yet implemented natively for %s", candidate.SourceTool)
	}
}

func buildTormentNexusDatabaseRecordInputs(ctx context.Context, db *sql.DB, candidate Candidate) ([]ImportedSessionRecordInput, error) {
	results := make([]ImportedSessionRecordInput, 0)
	seen := map[string]struct{}{}

	ledgerRows, _ := queryRowsAsMaps(ctx, db, `SELECT * FROM session_ledger ORDER BY datetime(created_at) DESC LIMIT ?`, 50)
	for _, row := range ledgerRows {
		id := stringValue(row["id"])
		if id == "" {
			continue
		}
		transcript := strings.TrimSpace(buildTormentNexusLedgerTranscript(row))
		if transcript == "" {
			continue
		}
		sourcePath := candidate.SourcePath + "#session_ledger:" + id
		if _, ok := seen[sourcePath]; ok {
			continue
		}
		seen[sourcePath] = struct{}{}
		lastModified := toTimestampMs(row["created_at"])
		externalSessionID := nonEmptyString(stringValue(row["conversation_id"]), id)
		title := nonEmptyString(sanitizeSentence(stringValue(row["summary"]), 120), "TormentNexus ledger "+id)
		metadata := map[string]any{
			"tormentnexusProject":         row["project"],
			"tormentnexusRole":            row["role"],
			"tormentnexusTable":           "session_ledger",
			"tormentnexusConversationId":  row["conversation_id"],
			"tormentnexusEventType":       nonEmptyString(stringValue(row["event_type"]), "session"),
			"tormentnexusConfidenceScore": row["confidence_score"],
			"tormentnexusImportance":      row["importance"],
		}
		if metadata["tormentnexusEventType"] == "correction" && intValueOr(row["importance"], 0) >= 3 && strings.TrimSpace(stringValue(row["summary"])) != "" {
			metadata["behavioralWarnings"] = []string{sanitizeSentence(stringValue(row["summary"]), 220)}
		}
		input, err := buildImportedSessionRecordInputFromTranscript(candidate.SourceTool, sourcePath, "tormentnexus-ledger", title, transcript, filepath.Dir(candidate.SourcePath), externalSessionID, lastModified, metadata, nil)
		if err != nil {
			return nil, err
		}
		results = append(results, input)
	}

	handoffRows, _ := queryRowsAsMaps(ctx, db, `SELECT * FROM session_handoffs ORDER BY datetime(COALESCE(updated_at, created_at)) DESC LIMIT ?`, 50)
	for _, row := range handoffRows {
		project := stringValue(row["project"])
		if project == "" {
			continue
		}
		transcript := strings.TrimSpace(buildTormentNexusHandoffTranscript(row))
		if transcript == "" {
			continue
		}
		sourcePath := candidate.SourcePath + "#session_handoffs:" + project
		if _, ok := seen[sourcePath]; ok {
			continue
		}
		seen[sourcePath] = struct{}{}
		parsedMetadata := parseJSONObject(stringValue(row["metadata"]))
		metadata := map[string]any{
			"tormentnexusProject": project,
			"tormentnexusTable":   "session_handoffs",
			"tormentnexusVersion": row["version"],
		}
		for key, value := range parsedMetadata {
			metadata[key] = value
		}
		lastModified := firstTimestamp(toTimestampMs(row["updated_at"]), toTimestampMs(row["created_at"]))
		workingDirectory := filepath.Dir(candidate.SourcePath)
		if extracted := extractWorkingDirectory(parsedMetadata); extracted != nil {
			workingDirectory = *extracted
		}
		input, err := buildImportedSessionRecordInputFromTranscript(candidate.SourceTool, sourcePath, "tormentnexus-handoff", "TormentNexus handoff "+project, transcript, workingDirectory, "handoff:"+project, lastModified, metadata, nil)
		if err != nil {
			return nil, err
		}
		results = append(results, input)
	}

	return results, nil
}

func buildLLMDatabaseRecordInputs(ctx context.Context, db *sql.DB, candidate Candidate) ([]ImportedSessionRecordInput, error) {
	hasResponses, err := tableExists(ctx, db, "responses")
	if err != nil {
		return nil, err
	}
	if !hasResponses {
		return nil, nil
	}
	hasConversations, _ := tableExists(ctx, db, "conversations")
	hasToolCalls, _ := tableExists(ctx, db, "tool_calls")
	hasToolResults, _ := tableExists(ctx, db, "tool_results")

	toolCallsByResponse := map[string][]string{}
	if hasToolCalls {
		rows, err := queryRowsAsMaps(ctx, db, `SELECT response_id, name FROM tool_calls ORDER BY id ASC`)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			responseID := stringValue(row["response_id"])
			name := strings.TrimSpace(stringValue(row["name"]))
			if responseID == "" || name == "" {
				continue
			}
			toolCallsByResponse[responseID] = append(toolCallsByResponse[responseID], name)
		}
	}
	toolResultsByResponse := map[string][]string{}
	if hasToolResults {
		rows, err := queryRowsAsMaps(ctx, db, `SELECT response_id, name FROM tool_results ORDER BY id ASC`)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			responseID := stringValue(row["response_id"])
			name := strings.TrimSpace(stringValue(row["name"]))
			if responseID == "" || name == "" {
				continue
			}
			toolResultsByResponse[responseID] = append(toolResultsByResponse[responseID], name)
		}
	}

	results := make([]ImportedSessionRecordInput, 0)
	seen := map[string]struct{}{}
	if hasConversations {
		conversationRows, err := queryRowsAsMaps(ctx, db, `
			SELECT c.id, c.name, c.model, MAX(r.datetime_utc) AS last_activity
			FROM conversations c
			JOIN responses r ON r.conversation_id = c.id
			GROUP BY c.id, c.name, c.model
			ORDER BY datetime(MAX(r.datetime_utc)) DESC
			LIMIT ?
		`, 50)
		if err != nil {
			return nil, err
		}
		for _, row := range conversationRows {
			conversationID := stringValue(row["id"])
			if conversationID == "" {
				continue
			}
			sourcePath := candidate.SourcePath + "#conversation:" + conversationID
			if _, ok := seen[sourcePath]; ok {
				continue
			}
			seen[sourcePath] = struct{}{}
			responseRows, err := queryRowsAsMaps(ctx, db, `
				SELECT id, model, prompt, system, prompt_json, response, response_json, datetime_utc, input_tokens, output_tokens, resolved_model
				FROM responses
				WHERE conversation_id = ?
				ORDER BY datetime(datetime_utc) ASC, id ASC
			`, conversationID)
			if err != nil {
				return nil, err
			}
			transcript := strings.TrimSpace(buildLLMConversationTranscript(responseRows, toolCallsByResponse, toolResultsByResponse))
			if transcript == "" {
				continue
			}
			inputTokens := sumNumericField(responseRows, "input_tokens")
			outputTokens := sumNumericField(responseRows, "output_tokens")
			metadata := map[string]any{
				"llmConversationId":    conversationID,
				"llmConversationModel": row["model"],
				"llmResponseCount":     len(responseRows),
				"llmInputTokens":       inputTokens,
				"llmOutputTokens":      outputTokens,
				"llmDatabasePath":      candidate.SourcePath,
			}
			title := nonEmptyString(sanitizeSentence(stringValue(row["name"]), 120), "llm conversation "+conversationID)
			input, err := buildImportedSessionRecordInputFromTranscript(candidate.SourceTool, sourcePath, "llm-conversation", title, transcript, filepath.Dir(candidate.SourcePath), conversationID, toTimestampMs(row["last_activity"]), metadata, nil)
			if err != nil {
				return nil, err
			}
			results = append(results, input)
		}
	}

	remaining := 50 - len(results)
	if remaining > 0 {
		orphanRows, err := queryRowsAsMaps(ctx, db, `
			SELECT id, model, prompt, system, prompt_json, response, response_json, datetime_utc, input_tokens, output_tokens, resolved_model
			FROM responses
			WHERE conversation_id IS NULL
			ORDER BY datetime(datetime_utc) DESC, id DESC
			LIMIT ?
		`, remaining)
		if err != nil {
			return nil, err
		}
		for _, row := range orphanRows {
			responseID := stringValue(row["id"])
			if responseID == "" {
				continue
			}
			sourcePath := candidate.SourcePath + "#response:" + responseID
			if _, ok := seen[sourcePath]; ok {
				continue
			}
			seen[sourcePath] = struct{}{}
			transcript := strings.TrimSpace(buildLLMConversationTranscript([]map[string]any{row}, toolCallsByResponse, toolResultsByResponse))
			if transcript == "" {
				continue
			}
			promptText := buildLLMLogText(row["prompt"], row["prompt_json"])
			metadata := map[string]any{
				"llmResponseId":    responseID,
				"llmModel":         row["model"],
				"llmResolvedModel": row["resolved_model"],
				"llmInputTokens":   intValueOr(row["input_tokens"], 0),
				"llmOutputTokens":  intValueOr(row["output_tokens"], 0),
				"llmDatabasePath":  candidate.SourcePath,
			}
			title := nonEmptyString(sanitizeSentence(promptText, 120), "llm response "+responseID)
			input, err := buildImportedSessionRecordInputFromTranscript(candidate.SourceTool, sourcePath, "llm-response", title, transcript, filepath.Dir(candidate.SourcePath), responseID, toTimestampMs(row["datetime_utc"]), metadata, nil)
			if err != nil {
				return nil, err
			}
			results = append(results, input)
		}
	}

	return results, nil
}

func buildImportedSessionRecordInputFromTranscript(sourceTool, sourcePath, sessionFormat, title, transcript, workingDirectory, externalSessionID string, lastModifiedAt *int64, metadata map[string]any, detectedModels []string) (ImportedSessionRecordInput, error) {
	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		return ImportedSessionRecordInput{}, fmt.Errorf("empty transcript after parsing")
	}
	transcriptHash := hashTranscript(transcript)
	lines := normalizedLines(transcript)
	now := time.Now().UTC().UnixMilli()
	excerpt := transcript
	if len(excerpt) > 320 {
		excerpt = excerpt[:320]
	}
	sessionID := "imported:" + sourceTool + ":" + transcriptHash[:12]
	memories := heuristicMemoryExtraction(transcript, sourceTool)
	retentionSummary := buildRetentionSummary(transcript, memories)
	for index := range memories {
		if memories[index].Metadata == nil {
			memories[index].Metadata = map[string]any{}
		}
		memories[index].Metadata["sourceTool"] = sourceTool
		memories[index].Metadata["path"] = sourcePath
		memories[index].Metadata["sessionId"] = sessionID
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["sourceTool"] = sourceTool
	metadata["sourcePath"] = sourcePath
	metadata["sessionFormat"] = sessionFormat
	metadata["contentLength"] = len(transcript)
	metadata["lineCount"] = len(lines)
	metadata["retentionSummary"] = retentionSummary
	if len(detectedModels) > 0 {
		metadata["detectedModels"] = detectedModels
	}
	return ImportedSessionRecordInput{
		SourceTool:        sourceTool,
		SourcePath:        sourcePath,
		ExternalSessionID: stringPtr(externalSessionID),
		Title:             stringPtr(title),
		SessionFormat:     sessionFormat,
		Transcript:        transcript,
		Excerpt:           stringPtr(excerpt),
		WorkingDirectory:  stringPtr(workingDirectory),
		TranscriptHash:    transcriptHash,
		NormalizedSession: map[string]any{
			"sessionId":         sessionID,
			"title":             title,
			"sourceTool":        sourceTool,
			"sourcePath":        sourcePath,
			"sessionFormat":     sessionFormat,
			"externalSessionId": externalSessionID,
			"transcriptHash":    transcriptHash,
			"lineCount":         len(lines),
			"contentLength":     len(transcript),
			"excerpt":           excerpt,
			"importedAt":        now,
			"lastModifiedAt":    nullableTimeValue(lastModifiedAt),
			"detectedModels":    detectedModels,
		},
		Metadata:       metadata,
		DiscoveredAt:   now,
		ImportedAt:     now,
		LastModifiedAt: lastModifiedAt,
		ParsedMemories: memories,
	}, nil
}

func queryRowsAsMaps(ctx context.Context, db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			row[column] = normalizeSQLValue(values[i])
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func normalizeSQLValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	default:
		return typed
	}
}

func tableExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var one int
	err := db.QueryRowContext(ctx, `SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ? LIMIT 1`, name).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func parseJSONStringArrayLoose(value any) []string {
	switch typed := value.(type) {
	case string:
		return parseJSONStringSlice(typed)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := strings.TrimSpace(stringValue(item)); text != "" {
				result = append(result, text)
			}
		}
		return uniqueStrings(result, 32)
	default:
		return []string{}
	}
}

func buildTormentNexusLedgerTranscript(row map[string]any) string {
	lines := []string{
		fmt.Sprintf("TormentNexus session ledger for project %s", sanitizeSentence(nonEmptyString(stringValue(row["project"]), "default"), 120)),
		stringValue(row["summary"]),
	}
	if eventType := stringValue(row["event_type"]); strings.TrimSpace(eventType) != "" && eventType != "session" {
		lines = append(lines, "Event type: "+eventType)
	}
	if confidence, ok := row["confidence_score"]; ok && stringValue(confidence) != "" {
		lines = append(lines, fmt.Sprintf("Confidence score: %v", confidence))
	}
	if importance := intValueOr(row["importance"], 0); importance > 0 {
		lines = append(lines, fmt.Sprintf("Importance: %d", importance))
	}
	if todos := parseJSONStringArrayLoose(row["todos"]); len(todos) > 0 {
		lines = append(lines, "Open TODOs:")
		for _, todo := range todos {
			lines = append(lines, "- "+todo)
		}
	}
	if decisions := parseJSONStringArrayLoose(row["decisions"]); len(decisions) > 0 {
		lines = append(lines, "Decisions:")
		for _, decision := range decisions {
			lines = append(lines, "- "+decision)
		}
	}
	if filesChanged := parseJSONStringArrayLoose(row["files_changed"]); len(filesChanged) > 0 {
		lines = append(lines, "Files changed:")
		for _, fileName := range filesChanged {
			lines = append(lines, "- "+fileName)
		}
	}
	if keywords := parseJSONStringArrayLoose(row["keywords"]); len(keywords) > 0 {
		lines = append(lines, "Keywords: "+strings.Join(keywords, ", "))
	}
	if createdAt := stringValue(row["created_at"]); createdAt != "" {
		lines = append(lines, "Created at: "+createdAt)
	}
	if role := stringValue(row["role"]); role != "" {
		lines = append(lines, "Role: "+role)
	}
	return strings.Join(filterNonEmpty(lines), "\n")
}

func buildTormentNexusHandoffTranscript(row map[string]any) string {
	lines := []string{
		fmt.Sprintf("TormentNexus handoff for project %s", sanitizeSentence(nonEmptyString(stringValue(row["project"]), "default"), 120)),
		stringValue(row["last_summary"]),
		stringValue(row["key_context"]),
	}
	if pendingTodo := parseJSONStringArrayLoose(row["pending_todo"]); len(pendingTodo) > 0 {
		lines = append(lines, "Open TODOs:")
		for _, todo := range pendingTodo {
			lines = append(lines, "- "+todo)
		}
	}
	if activeDecisions := parseJSONStringArrayLoose(row["active_decisions"]); len(activeDecisions) > 0 {
		lines = append(lines, "Active decisions:")
		for _, decision := range activeDecisions {
			lines = append(lines, "- "+decision)
		}
	}
	if keywords := parseJSONStringArrayLoose(row["keywords"]); len(keywords) > 0 {
		lines = append(lines, "Keywords: "+strings.Join(keywords, ", "))
	}
	if activeBranch := stringValue(row["active_branch"]); activeBranch != "" {
		lines = append(lines, "Active branch: "+activeBranch)
	}
	if version := intValueOr(row["version"], 0); version > 0 {
		lines = append(lines, fmt.Sprintf("Version: %d", version))
	}
	if updatedAt := stringValue(row["updated_at"]); updatedAt != "" {
		lines = append(lines, "Updated at: "+updatedAt)
	}
	return strings.Join(filterNonEmpty(lines), "\n")
}

func buildLLMLogText(primary any, jsonFallback any) string {
	if text := strings.TrimSpace(stringValue(primary)); text != "" {
		return text
	}
	fragments := uniqueStrings(extractJSONTextFragments(parseJSONValue(jsonFallback)), 40)
	return strings.TrimSpace(strings.Join(fragments, "\n"))
}

func buildLLMConversationTranscript(rows []map[string]any, toolCallsByResponse map[string][]string, toolResultsByResponse map[string][]string) string {
	entries := make([]string, 0, len(rows))
	for _, row := range rows {
		responseID := stringValue(row["id"])
		systemText := buildLLMLogText(row["system"], nil)
		promptText := buildLLMLogText(row["prompt"], row["prompt_json"])
		responseText := buildLLMLogText(row["response"], row["response_json"])
		toolCalls := toolCallsByResponse[responseID]
		toolResults := toolResultsByResponse[responseID]
		blocks := []string{}
		if systemText != "" {
			blocks = append(blocks, "System: "+systemText)
		}
		if promptText != "" {
			blocks = append(blocks, "User: "+promptText)
		}
		assistantFragments := []string{}
		if responseText != "" {
			assistantFragments = append(assistantFragments, responseText)
		}
		for _, name := range toolCalls {
			assistantFragments = append(assistantFragments, "[Tool Call: "+name+"]")
		}
		for _, name := range toolResults {
			assistantFragments = append(assistantFragments, "[Tool Result: "+name+"]")
		}
		if len(assistantFragments) > 0 {
			blocks = append(blocks, "Assistant: "+strings.Join(assistantFragments, "\n"))
		}
		if len(blocks) > 0 {
			entries = append(entries, strings.Join(blocks, "\n\n"))
		}
	}
	return strings.Join(entries, "\n\n")
}

func parseJSONValue(value any) any {
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return value
	}
	var parsed any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return value
	}
	return parsed
}

func extractWorkingDirectory(metadata map[string]any) *string {
	for _, key := range []string{"cwd", "workingDirectory", "workspacePath", "repoPath", "projectPath"} {
		if value := strings.TrimSpace(stringValue(metadata[key])); value != "" {
			return &value
		}
	}
	return nil
}

func toTimestampMs(value any) *int64 {
	switch typed := value.(type) {
	case int64:
		v := typed
		if v < 1_000_000_000_000 {
			v *= 1000
		}
		return &v
	case int:
		v := int64(typed)
		if v < 1_000_000_000_000 {
			v *= 1000
		}
		return &v
	case float64:
		v := int64(typed)
		if v < 1_000_000_000_000 {
			v *= 1000
		}
		return &v
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil
		}
		if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
			v := parsed.UTC().UnixMilli()
			return &v
		}
	}
	return nil
}

func intValueOr(value any, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		return fallback
	default:
		return fallback
	}
}

func sumNumericField(rows []map[string]any, key string) int {
	total := 0
	for _, row := range rows {
		total += intValueOr(row[key], 0)
	}
	return total
}

func nonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func filterNonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func firstTimestamp(values ...*int64) *int64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
