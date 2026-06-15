//go:build ignore
// +build ignore

package tools

/**
 * @file semgrep.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Semgrep static analysis security scanning.
 * Replaces `semgrep` (STDIO) and `semgrepstream` (SSE) entries in mcp.json.
 *
 * Uses the local `semgrep` binary AND the Semgrep Cloud Platform API.
 * Improvements over original:
 *  - Unified tool: runs semgrep CLI locally or calls the cloud API.
 *  - Returns structured JSON findings with severity/rule metadata.
 *  - Context-aware with timeout.
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const semgrepAPIBase = "https://semgrep.dev/api/v1"

func semgrepToken() string {
	return os.Getenv("SEMGREP_APP_TOKEN")
}

// HandleSemgrepScan runs semgrep locally on a target path.
// Tool: semgrep_scan
func HandleSemgrepScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ := getString(args, "target", "path", "directory")
	if target == "" {
		target = "."
	}

	// Resolve to absolute
	absTarget, e := filepath.Abs(target)
	if e != nil {
		return err(fmt.Sprintf("invalid target path: %v", e))
	}

	// Check semgrep is available
	semgrepPath, e := exec.LookPath("semgrep")
	if e != nil {
		return err("semgrep binary not found in PATH. Please install semgrep (pip install semgrep or https://semgrep.dev/docs/getting-started)")
	}

	cmdArgs := []string{"--json"}

	// Config/ruleset
	if config, _ := getString(args, "config", "ruleset"); config != "" {
		cmdArgs = append(cmdArgs, "--config", config)
	} else {
		cmdArgs = append(cmdArgs, "--config", "auto")
	}

	// Severity filter
	if severity, _ := getString(args, "severity"); severity != "" {
		cmdArgs = append(cmdArgs, "--severity", strings.ToUpper(severity))
	}

	// Language filter
	if lang, _ := getString(args, "lang", "language"); lang != "" {
		cmdArgs = append(cmdArgs, "--lang", lang)
	}

	// Max findings
	maxFindings := getInt(args, "max_findings", "limit")
	if maxFindings > 0 {
		cmdArgs = append(cmdArgs, "--max-target-bytes", fmt.Sprintf("%d", maxFindings*1000))
	}

	cmdArgs = append(cmdArgs, absTarget)

	cmd := exec.CommandContext(ctx, semgrepPath, cmdArgs...)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// semgrep exits 1 when findings are found, 0 when clean
	runErr := cmd.Run()
	_ = runErr

	stdoutStr := stdout.String()
	if stdoutStr == "" {
		if stderr.Len() > 0 {
			return err("semgrep error: " + stderr.String())
		}
		return ok("No findings.")
	}

	// Parse JSON output
	var result map[string]interface{}
	if e := json.Unmarshal([]byte(stdoutStr), &result); e != nil {
		return ok(stdoutStr) // Return raw output on parse failure
	}

	findings, _ := result["results"].([]interface{})
	if len(findings) == 0 {
		return ok(fmt.Sprintf("✓ No security findings in %s", absTarget))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Semgrep found %d finding(s) in %s:\n\n%s", len(findings), absTarget, string(out)))
}

// HandleSemgrepCloudScan submits a scan via Semgrep Cloud Platform API.
// Tool: semgrep_cloud_scan
func HandleSemgrepCloudScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := semgrepToken()
	if token == "" {
		return err("SEMGREP_APP_TOKEN environment variable is required for cloud scanning")
	}

	// Get recent scan results from cloud
	req, e := http.NewRequestWithContext(ctx, "GET", semgrepAPIBase+"/findings", nil)
	if e != nil {
		return err(fmt.Sprintf("Failed to create request: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Query params
	q := req.URL.Query()
	if depl, _ := getString(args, "deployment_slug"); depl != "" {
		q.Set("deployment_slug", depl)
	}
	if severity, _ := getString(args, "severity"); severity != "" {
		q.Set("severity", severity)
	}
	if repoName, _ := getString(args, "repo_name"); repoName != "" {
		q.Set("repo_name", repoName)
	}
	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 25
	}
	q.Set("page_size", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Semgrep cloud API request failed: %v", e))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Semgrep Cloud API error (HTTP %d): %s", resp.StatusCode, string(body)))
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(string(body))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleSemgrepRuleSearch searches for community Semgrep rules.
// Tool: semgrep_search_rules
func HandleSemgrepRuleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	lang, _ := getString(args, "lang", "language")
	category, _ := getString(args, "category")

	searchURL := "https://semgrep.dev/api/registry/rules"

	req, e := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if e != nil {
		return err(fmt.Sprintf("Failed to create request: %v", e))
	}

	q := req.URL.Query()
	if query != "" {
		q.Set("q", query)
	}
	if lang != "" {
		q.Set("lang", lang)
	}
	if category != "" {
		q.Set("category", category)
	}
	q.Set("per_page", "20")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 20 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Semgrep registry request failed: %v", e))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("Semgrep registry error (HTTP %d): %s", resp.StatusCode, string(body)))
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(string(body))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
