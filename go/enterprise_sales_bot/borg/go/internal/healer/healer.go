// Package healer provides self-healing diagnosis and auto-fix capabilities
// ported from packages/core/src/services/HealerService.ts.
//
// It uses an LLM to analyze errors, diagnose root causes, and generate fixes.
// Supports: error analysis, fix generation, auto-heal, and heal history.
package healer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
	"github.com/tormentnexushq/tormentnexus-go/internal/codeexec"
	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

// Diagnosis represents an LLM-generated diagnosis of an error.
type Diagnosis struct {
	ErrorType    string  `json:"errorType"`
	Description  string  `json:"description"`
	File         string  `json:"file,omitempty"`
	Line         int     `json:"line,omitempty"`
	SuggestedFix string  `json:"suggestedFix"`
	Confidence   float64 `json:"confidence"`
}

// FixPlan represents a plan to fix a diagnosed error.
type FixPlan struct {
	ID            string             `json:"id"`
	Diagnosis     Diagnosis          `json:"diagnosis"`
	FilesToModify []FileModification `json:"filesToModify"`
	Explanation   string             `json:"explanation"`
}

// FileModification represents a file to be modified by a fix.
type FileModification struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// HealRecord tracks a single heal attempt.
type HealRecord struct {
	Timestamp int64   `json:"timestamp"`
	Error     string  `json:"error"`
	Fix       FixPlan `json:"fix"`
	Success   bool    `json:"success"`
	Attempts  int     `json:"attempts"`
}

// HealerService provides self-healing capabilities.
type HealerService struct {
	mu       sync.RWMutex
	history  []HealRecord
	provider ai.Provider
	model    string
	executor *codeexec.CodeExecutor
	vault    controlplane.MemoryVault
	onHeal   func(HealRecord)
}

// NewHealerService creates a new healer with the given LLM provider, executor and vault.
func NewHealerService(provider ai.Provider, model string, executor *codeexec.CodeExecutor, vault controlplane.MemoryVault) *HealerService {
	if model == "" {
		model = "openrouter/auto"
	}
	if executor == nil {
		executor = codeexec.NewCodeExecutor()
	}
	return &HealerService{
		provider: provider,
		model:    model,
		executor: executor,
		vault:    vault,
	}
}

// OnHeal registers a callback for heal events.
func (hs *HealerService) OnHeal(fn func(HealRecord)) {
	hs.onHeal = fn
}

// AnalyzeError uses an LLM to diagnose an error.
func (hs *HealerService) AnalyzeError(ctx context.Context, errorStr string, contextStr string) (*Diagnosis, error) {
	if contextStr == "" {
		contextStr = "No additional context."
	}

	prompt := fmt.Sprintf(`You are The Healer, an expert debugging agent.
Analyze the following error and context.
Provide a diagnosis and a suggested fix.

Error:
%s

Context:
%s

Return JSON format:
{
    "errorType": "SyntaxError|RuntimeError|LogicError|...",
    "description": "Short explanation",
    "file": "path/to/culprit.ts (if known)",
    "line": 123 (if known),
    "suggestedFix": "Code snippet or description of fix",
    "confidence": 0.0 to 1.0
}`, errorStr, contextStr)

	if hs.provider == nil {
		return &Diagnosis{
			ErrorType:    "Unknown",
			Description:  errorStr,
			SuggestedFix: "Manual review required (no LLM provider)",
			Confidence:   0,
		}, nil
	}

	resp, err := hs.provider.GenerateText(ctx, hs.model, []ai.Message{
		{Role: "system", Content: "You are a JSON-only debugging tool."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("LLM diagnosis failed: %w", err)
	}

	var diag Diagnosis
	text := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(text), &diag); err != nil {
		return &Diagnosis{
			ErrorType:    "Unknown",
			Description:  "Failed to parse LLM diagnosis",
			SuggestedFix: "Manual review required",
			Confidence:   0,
		}, nil
	}

	return &diag, nil
}

// GenerateFix creates a fix plan for the given diagnosis.
func (hs *HealerService) GenerateFix(ctx context.Context, diag *Diagnosis) (*FixPlan, error) {
	if diag.File == "" {
		return nil, fmt.Errorf("cannot generate fix without file path")
	}

	// Read file content
	content, err := os.ReadFile(diag.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", diag.File, err)
	}

	if hs.provider == nil {
		return &FixPlan{
			ID:          fmt.Sprintf("fix_%d", time.Now().UnixMilli()),
			Diagnosis:   *diag,
			Explanation: "No LLM provider available for fix generation",
		}, nil
	}

	prompt := fmt.Sprintf(`You are The Healer.
Generate a fix for the following file based on the diagnosis.

Diagnosis: %s
Suggested Fix: %s

File Content:
%s

Return JSON format:
{
    "explanation": "Why this fix works",
    "newContent": "The entire new file content"
}`, diag.Description, diag.SuggestedFix, string(content))

	resp, err := hs.provider.GenerateText(ctx, hs.model, []ai.Message{
		{Role: "system", Content: "You are a code repair agent. Return only JSON with 'explanation' and 'newContent'."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("LLM fix generation failed: %w", err)
	}

	var result struct {
		Explanation string `json:"explanation"`
		NewContent  string `json:"newContent"`
	}
	text := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse fix plan")
	}

	return &FixPlan{
		ID:        fmt.Sprintf("fix_%d", time.Now().UnixMilli()),
		Diagnosis: *diag,
		FilesToModify: []FileModification{
			{Path: diag.File, Content: result.NewContent},
		},
		Explanation: result.Explanation,
	}, nil
}

// ApplyFix writes the fix plan's files to disk.
func (hs *HealerService) ApplyFix(plan *FixPlan) error {
	for _, fm := range plan.FilesToModify {
		dir := filepath.Dir(fm.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(fm.Path, []byte(fm.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fm.Path, err)
		}
	}
	return nil
}

// HealAndVerify performs the core autonomous loop: diagnose -> fix -> verify -> retry.
func (hs *HealerService) HealAndVerify(ctx context.Context, errorStr string, contextStr string, maxAttempts int) (bool, error) {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	currentError := errorStr
	currentContext := contextStr
	attempts := 0

	var lastPlan *FixPlan

	for attempts < maxAttempts {
		attempts++
		diag, err := hs.AnalyzeError(ctx, currentError, currentContext)
		if err != nil {
			return false, err
		}

		if diag.Confidence < 0.6 {
			return false, fmt.Errorf("confidence too low for auto-heal (%.2f)", diag.Confidence)
		}

		if diag.File == "" {
			return false, fmt.Errorf("no file identified to fix")
		}

		plan, err := hs.GenerateFix(ctx, diag)
		if err != nil {
			return false, err
		}
		lastPlan = plan

		if err := hs.ApplyFix(plan); err != nil {
			hs.recordHeal(ctx, currentError, *plan, false, attempts)
			return false, err
		}

		// Verify fix
		vErr := hs.verifyFix(ctx, diag.File)
		if vErr == nil {
			hs.recordHeal(ctx, errorStr, *plan, true, attempts)
			return true, nil
		}

		// Update context for next attempt
		currentError = vErr.Error()
		currentContext = fmt.Sprintf("Attempted fix: %s. But verification failed with: %s", plan.Explanation, currentError)
	}

	if lastPlan != nil {
		hs.recordHeal(ctx, errorStr, *lastPlan, false, attempts)
	}
	return false, fmt.Errorf("max attempts reached without successful fix")
}

// verifyFix runs relevant tests or type checks to verify the fix.
func (hs *HealerService) verifyFix(ctx context.Context, culpritFile string) error {
	ext := filepath.Ext(culpritFile)

	var commands []string
	if ext == ".ts" {
		testFile := strings.TrimSuffix(culpritFile, ".ts") + ".test.ts"
		if _, err := os.Stat(testFile); err == nil {
			commands = append(commands, fmt.Sprintf("npx vitest run %s", testFile))
		} else {
			commands = append(commands, fmt.Sprintf("npx tsc --noEmit %s", culpritFile))
		}
	} else if ext == ".go" {
		dir := filepath.Dir(culpritFile)
		commands = append(commands, fmt.Sprintf("go test -v %s", dir))
	}

	for _, cmdStr := range commands {
		res, err := hs.executor.Execute(ctx, codeexec.ExecutionConfig{
			Language: codeexec.Shell,
			Code:     cmdStr,
		})
		if err != nil {
			return err
		}
		if res.ExitCode != 0 {
			return fmt.Errorf("verification command failed (%s): %s", cmdStr, res.Stderr)
		}
	}

	return nil
}

// Heal performs a single heal cycle (deprecated in favor of HealAndVerify).
func (hs *HealerService) Heal(ctx context.Context, errorStr string, contextStr string) (bool, error) {
	return hs.HealAndVerify(ctx, errorStr, contextStr, 1)
}

// AutoHeal performs a one-shot auto-heal on an error log string.
func (hs *HealerService) AutoHeal(ctx context.Context, errorLog string) (*HealResult, error) {
	diag, err := hs.AnalyzeError(ctx, errorLog, "")
	if err != nil {
		return nil, err
	}

	if diag.File == "" || diag.SuggestedFix == "" {
		return &HealResult{Success: false, Diagnosis: diag}, nil
	}

	// Resolve file path
	filePath := diag.File
	if !filepath.IsAbs(filePath) {
		if cwd, err := os.Getwd(); err == nil {
			filePath = filepath.Join(cwd, filePath)
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &HealResult{Success: false, Diagnosis: diag, File: filePath}, nil
	}

	diag.File = filePath
	plan, err := hs.GenerateFix(ctx, diag)
	if err != nil {
		return nil, err
	}

	if err := hs.ApplyFix(plan); err != nil {
		hs.recordHeal(ctx, errorLog, *plan, false, 1)
		return &HealResult{Success: false, File: filePath, Fix: diag.SuggestedFix}, err
	}

	hs.recordHeal(ctx, errorLog, *plan, true, 1)
	return &HealResult{Success: true, File: filePath, Fix: diag.SuggestedFix}, nil
}

// HealResult is the result of an auto-heal attempt.
type HealResult struct {
	Success   bool       `json:"success"`
	File      string     `json:"file,omitempty"`
	Fix       string     `json:"fix,omitempty"`
	Diagnosis *Diagnosis `json:"diagnosis,omitempty"`
}

// GetHistory returns all heal records.
func (hs *HealerService) GetHistory() []HealRecord {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	result := make([]HealRecord, len(hs.history))
	copy(result, hs.history)
	return result
}

func (hs *HealerService) recordHeal(ctx context.Context, errorStr string, plan FixPlan, success bool, attempts int) {
	record := HealRecord{
		Timestamp: time.Now().UnixMilli(),
		Error:     errorStr,
		Fix:       plan,
		Success:   success,
		Attempts:  attempts,
	}

	hs.mu.Lock()
	hs.history = append(hs.history, record)
	hs.mu.Unlock()

	// Persist to L2 Vault
	if hs.vault != nil {
		content, _ := json.Marshal(record)
		entry := controlplane.L2VaultRecord{
			ID:         fmt.Sprintf("heal-%s-%d", plan.ID, record.Timestamp),
			SessionID:  "kernel-healer",
			Type:       controlplane.MemoryLongTerm,
			Content:    fmt.Sprintf("Heal Event: %s (Success: %v, Attempts: %d)\nFix: %s\nDetails: %s", plan.Diagnosis.Description, success, attempts, plan.Explanation, string(content)),
			Importance: 0.8,
			HeatScore:  100.0,
			CreatedAt:  time.Now(),
		}
		_ = hs.vault.Commit(ctx, entry)
	}

	if hs.onHeal != nil {
		hs.onHeal(record)
	}
}

// --- helpers ---

var jsonBlockRe = regexp.MustCompile("(?s)```(?:json)?\\s*?(.*?)\\s*?```")

// extractJSON extracts a JSON object from text, handling markdown fences.
func extractJSON(text string) string {
	// Try fenced code block first
	if match := jsonBlockRe.FindStringSubmatch(text); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	// Try finding raw braces
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(text[start : end+1])
	}

	return strings.TrimSpace(text)
}
