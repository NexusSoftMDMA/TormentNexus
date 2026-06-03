package sessionimport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BuildManifest creates a manifest from validated candidates
func BuildManifest(candidates []ValidationResult) Manifest {
	return Manifest{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Count:       len(candidates),
		Candidates:  candidates,
	}
}

// ExportRecord represents a single exported session record
type ExportRecord struct {
	SessionID   string    `json:"sessionId"`
	SourceTool  string    `json:"sourceTool"`
	SourcePath  string    `json:"sourcePath"`
	ExportedAt  time.Time `json:"exportedAt"`
	Format      string    `json:"format"`
	ContentSize int64     `json:"contentSize"`
}

// ExportResult represents the outcome of an export operation
type ExportResult struct {
	OutputPath     string         `json:"outputPath"`
	ExportedCount  int            `json:"exportedCount"`
	SkippedCount   int            `json:"skippedCount"`
	TotalSize      int64          `json:"totalSize"`
	ExportedAt     time.Time      `json:"exportedAt"`
	Records        []ExportRecord `json:"records"`
}

// ExportSessions scans for session candidates, validates them, and writes a unified export
func ExportSessions(workspaceRoot, homeDir, outputDir string) (*ExportResult, error) {
	scanner := NewScanner(workspaceRoot, homeDir, 200)
	candidates, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	if outputDir == "" {
		outputDir = filepath.Join(workspaceRoot, ".tormentnexus", "exports")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	result := &ExportResult{
		OutputPath: outputDir,
		ExportedAt: time.Now().UTC(),
		Records:    make([]ExportRecord, 0, len(candidates)),
	}

	for i, candidate := range candidates {
		// Read the source file
		content, err := os.ReadFile(candidate.SourcePath)
		if err != nil {
			result.SkippedCount++
			continue
		}

		sessionID := fmt.Sprintf("%s-%d-%d", candidate.SourceTool, time.Now().UnixMilli(), i)
		exportPath := filepath.Join(outputDir, sessionID+"."+candidate.SessionFormat)

		if err := os.WriteFile(exportPath, content, 0644); err != nil {
			result.SkippedCount++
			continue
		}

		record := ExportRecord{
			SessionID:   sessionID,
			SourceTool:  candidate.SourceTool,
			SourcePath:  candidate.SourcePath,
			ExportedAt:  time.Now().UTC(),
			Format:      candidate.SessionFormat,
			ContentSize: int64(len(content)),
		}

		result.Records = append(result.Records, record)
		result.ExportedCount++
		result.TotalSize += record.ContentSize
	}

	// Write manifest
	manifest := struct {
		ExportResult
		Manifest Manifest `json:"manifest"`
	}{
		ExportResult: *result,
	}

	validated := make([]ValidationResult, 0, len(candidates))
	for _, c := range candidates {
		validated = append(validated, ValidateCandidate(c))
	}
	manifest.Manifest = BuildManifest(validated)

	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	manifestPath := filepath.Join(outputDir, "export-manifest.json")
	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		return result, fmt.Errorf("failed to write manifest: %w", err)
	}

	return result, nil
}
