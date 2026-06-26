package sessionimport

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

var knownModelHints = []string{
	"gpt-4", "gpt-5", "claude", "gemini", "codex", "o1", "o3", "o4",
}

func ValidateCandidate(candidate Candidate) ValidationResult {
	result := ValidationResult{
		SourceTool:     candidate.SourceTool,
		SourceType:     detectSourceType(candidate),
		SourcePath:     candidate.SourcePath,
		Format:         candidate.SessionFormat,
		LastModifiedAt: candidate.LastModifiedAt,
		EstimatedSize:  candidate.EstimatedSize,
		DetectedModels: []string{},
	}

	content, err := os.ReadFile(candidate.SourcePath)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	lowerContent := strings.ToLower(string(content))
	for _, hint := range knownModelHints {
		if strings.Contains(lowerContent, hint) {
			result.DetectedModels = append(result.DetectedModels, hint)
		}
	}
	result.DetectedModels = dedupeStrings(result.DetectedModels)

	switch candidate.SessionFormat {
	case "json":
		if !json.Valid(content) {
			result.Errors = append(result.Errors, "invalid JSON document")
			return result
		}
	case "jsonl":
		if err := validateJSONL(content); err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result
		}
	default:
		if len(bytes.TrimSpace(content)) == 0 {
			result.Errors = append(result.Errors, "empty session artifact")
			return result
		}
	}

	result.Valid = true
	return result
}

func validateJSONL(content []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNumber := 0
	validLineCount := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !json.Valid([]byte(line)) {
			return newValidationError("invalid JSONL line", lineNumber)
		}
		validLineCount++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if validLineCount == 0 {
		return newValidationError("empty JSONL document", 0)
	}
	return nil
}

func detectSourceType(candidate Candidate) string {
	lowerPath := strings.ToLower(candidate.SourcePath)
	switch {
	case strings.HasSuffix(lowerPath, ".db"):
		return "database-log"
	case strings.Contains(lowerPath, "checkpoint"):
		return "checkpoint"
	case strings.Contains(lowerPath, "conversation"):
		return "conversation-export"
	case strings.Contains(lowerPath, "history"):
		return "history-log"
	case strings.Contains(lowerPath, "session"):
		return "session"
	default:
		return "artifact"
	}
}

func dedupeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || slices.Contains(result, value) {
			continue
		}
		result = append(result, value)
	}
	return result
}

type validationError struct {
	message string
	line    int
}

func newValidationError(message string, line int) error {
	return validationError{message: message, line: line}
}

func (e validationError) Error() string {
	if e.line <= 0 {
		return e.message
	}
	return e.message + " at line " + strconv.Itoa(e.line)
}

func DetectFormatFromPath(path string) string {
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if extension == "" {
		return "text"
	}
	return extension
}
