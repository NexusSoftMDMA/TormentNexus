package sessionimport

/**
 * @file transcript.go
 * @module go/internal/sessionimport
 *
 * WHAT: Transcript extraction from discovered session files.
 *       Parses JSON, JSONL, and Markdown session files to extract
 *       conversation transcripts, titles, and metadata.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// TranscriptResult holds the extracted transcript and metadata from a session file.
type TranscriptResult struct {
	SourceTool    string            `json:"sourceTool"`
	SourcePath    string            `json:"sourcePath"`
	Title         string            `json:"title,omitempty"`
	WorkingDir    string            `json:"workingDir,omitempty"`
	ExternalID    string            `json:"externalId,omitempty"`
	Transcript    string            `json:"transcript"`
	MessageCount  int               `json:"messageCount"`
	Participants  []string          `json:"participants,omitempty"`
	Language      string            `json:"language,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	ExtractedAt   time.Time         `json:"extractedAt"`
}

// ExtractTranscript reads a session file and extracts its transcript.
func ExtractTranscript(candidate Candidate) (*TranscriptResult, error) {
	data, err := os.ReadFile(candidate.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Limit to 2MB
	if len(data) > 2*1024*1024 {
		data = data[:2*1024*1024]
	}

	result := &TranscriptResult{
		SourceTool:  candidate.SourceTool,
		SourcePath:  candidate.SourcePath,
		ExtractedAt: time.Now().UTC(),
		Metadata:    make(map[string]string),
	}

	switch candidate.SessionFormat {
	case "json":
		extractJSONTranscript(result, data)
	case "jsonl":
		extractJSONLTranscript(result, data)
	case "md", "mdx", "txt", "log":
		extractTextTranscript(result, data)
	default:
		result.Transcript = string(data)
	}

	return result, nil
}

func extractJSONTranscript(result *TranscriptResult, data []byte) {
	// Try as object
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		extractFromJSONObject(result, obj)
		return
	}

	// Try as array
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		extractFromJSONArray(result, arr)
		return
	}

	// Fallback: just use raw text
	result.Transcript = string(data)
}

func extractFromJSONObject(result *TranscriptResult, obj map[string]interface{}) {
	// Title
	for _, key := range []string{"title", "name", "sessionName", "conversationName", "summary"} {
		if v, ok := obj[key].(string); ok && v != "" {
			result.Title = v
			break
		}
	}

	// Working directory
	for _, key := range []string{"workingDirectory", "cwd", "workDir", "projectPath"} {
		if v, ok := obj[key].(string); ok && v != "" {
			result.WorkingDir = v
			break
		}
	}

	// External ID
	for _, key := range []string{"id", "sessionId", "uuid", "threadId"} {
		if v, ok := obj[key].(string); ok && v != "" {
			result.ExternalID = v
			break
		}
	}

	// Messages
	extractMessages(result, obj["messages"])
	extractMessages(result, obj["conversation"])
	extractMessages(result, obj["turns"])
}

func extractFromJSONArray(result *TranscriptResult, arr []map[string]interface{}) {
	var parts []string
	seen := make(map[string]bool)

	for _, entry := range arr {
		role, _ := entry["role"].(string)
		content, _ := entry["content"].(string)
		if content == "" {
			content, _ = entry["text"].(string)
		}

		if role != "" && content != "" {
			parts = append(parts, fmt.Sprintf("[%s] %s", role, content))
			if !seen[role] {
				seen[role] = true
				result.Participants = append(result.Participants, role)
			}
		}
	}

	result.MessageCount = len(parts)
	result.Transcript = strings.Join(parts, "\n\n")

	if result.Title == "" && len(arr) > 0 {
		extractFromJSONObject(result, arr[0])
	}
}

func extractMessages(result *TranscriptResult, raw interface{}) {
	if raw == nil {
		return
	}

	switch v := raw.(type) {
	case []interface{}:
		var parts []string
		seen := make(map[string]bool)
		for _, item := range v {
			msg, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)
			if content == "" {
				content, _ = msg["text"].(string)
			}
			if role != "" && content != "" {
				parts = append(parts, fmt.Sprintf("[%s] %s", role, content))
				if !seen[role] {
					seen[role] = true
					result.Participants = append(result.Participants, role)
				}
			}
		}
		if len(parts) > 0 {
			result.MessageCount = len(parts)
			result.Transcript = strings.Join(parts, "\n\n")
		}

	case map[string]interface{}:
		// Might be nested
		extractFromJSONObject(result, v)
	}
}

func extractJSONLTranscript(result *TranscriptResult, data []byte) {
	var parts []string
	seen := make(map[string]bool)
	firstObj := true

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		// Extract metadata from first entry
		if firstObj {
			extractFromJSONObject(result, entry)
			firstObj = false
		}

		role, _ := entry["role"].(string)
		content, _ := entry["content"].(string)
		if content == "" {
			content, _ = entry["text"].(string)
		}

		if role != "" && content != "" {
			parts = append(parts, fmt.Sprintf("[%s] %s", role, content))
			if !seen[role] {
				seen[role] = true
				result.Participants = append(result.Participants, role)
			}
		}
	}

	result.MessageCount = len(parts)
	if len(parts) > 0 {
		result.Transcript = strings.Join(parts, "\n\n")
	}
}

func extractTextTranscript(result *TranscriptResult, data []byte) {
	content := string(data)

	// Try to extract title from first heading
	headingRe := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	if matches := headingRe.FindStringSubmatch(content); len(matches) > 1 {
		result.Title = matches[1]
	}

	// Count message-like patterns
	result.MessageCount = strings.Count(content, "\n## ") + strings.Count(content, "\n### ")

	// Use content as transcript (truncated)
	maxLen := 50000
	if len(content) > maxLen {
		result.Transcript = content[:maxLen]
	} else {
		result.Transcript = content
	}
}
