//go:build ignore
// +build ignore

package tools

/**
 * @file ast_grep.go
 * @module go/internal/tools
 *
 * WHAT: Go-native reimplementation of ast-grep-mcp tools.
 * Exposes ast-grep capabilities natively through Go exec.Command wrapper.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// HandleDumpSyntaxTree dumps the syntax tree of a specific code snippet or file.
func HandleDumpSyntaxTree(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := getString(args, "code")
	filePath, _ := getString(args, "file_path", "filePath", "path")
	language, _ := getString(args, "language", "lang")

	if code == "" && filePath == "" {
		return err("either code or file_path parameter is required")
	}

	// Verify sg is installed
	if _, errPath := exec.LookPath("sg"); errPath != nil {
		return err("ast-grep command 'sg' not found in PATH. Please install ast-grep (npm i -g @ast-grep/cli or cargo install ast-grep) to use this tool.")
	}

	var cmd *exec.Cmd
	var tempFile string

	if code != "" {
		// Write code to a temporary file to dump syntax tree
		ext := ".txt"
		if language != "" {
			ext = "." + language
		}
		tmpFile, errTmp := os.CreateTemp("", "ast-grep-dump-*"+ext)
		if errTmp != nil {
			return err(fmt.Sprintf("failed to create temporary file: %v", errTmp))
		}
		tempFile = tmpFile.Name()
		defer os.Remove(tempFile)

		if _, errWrite := tmpFile.Write([]byte(code)); errWrite != nil {
			tmpFile.Close()
			return err(fmt.Sprintf("failed to write temporary file: %v", errWrite))
		}
		tmpFile.Close()
		cmd = exec.CommandContext(ctx, "sg", "run", "--pattern", "$$$", tempFile, "--json")
	} else {
		cmd = exec.CommandContext(ctx, "sg", "run", "--pattern", "$$$", filePath, "--json")
	}

	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("ast-grep failed: %v\nOutput: %s", runErr, string(output)))
	}

	return ok(string(output))
}

// HandleTestMatchCodeRule tests a YAML ast-grep rule against a code snippet.
func HandleTestMatchCodeRule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ruleYaml, _ := getString(args, "rule", "rule_yaml", "yaml")
	code, _ := getString(args, "code")
	language, _ := getString(args, "language", "lang")

	if ruleYaml == "" {
		return err("rule parameter (YAML rule string) is required")
	}
	if code == "" {
		return err("code snippet parameter is required")
	}

	if _, errPath := exec.LookPath("sg"); errPath != nil {
		return err("ast-grep command 'sg' not found in PATH.")
	}

	// Write rule to a temp file
	tmpRule, errTmp := os.CreateTemp("", "ast-grep-rule-*.yaml")
	if errTmp != nil {
		return err(fmt.Sprintf("failed to create temporary rule file: %v", errTmp))
	}
	rulePath := tmpRule.Name()
	defer os.Remove(rulePath)

	if _, errWrite := tmpRule.Write([]byte(ruleYaml)); errWrite != nil {
		tmpRule.Close()
		return err(fmt.Sprintf("failed to write temporary rule file: %v", errWrite))
	}
	tmpRule.Close()

	// Write code snippet to a temp file
	ext := ".txt"
	if language != "" {
		ext = "." + language
	}
	tmpCode, errTmpCode := os.CreateTemp("", "ast-grep-code-*"+ext)
	if errTmpCode != nil {
		return err(fmt.Sprintf("failed to create temporary code file: %v", errTmpCode))
	}
	codePath := tmpCode.Name()
	defer os.Remove(codePath)

	if _, errWrite := tmpCode.Write([]byte(code)); errWrite != nil {
		tmpCode.Close()
		return err(fmt.Sprintf("failed to write temporary code file: %v", errWrite))
	}
	tmpCode.Close()

	// Run sg scan with the temp rule
	cmd := exec.CommandContext(ctx, "sg", "scan", "--rule", rulePath, codePath, "--json")
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		// ast-grep scan exits with non-zero if issues are found, which is normal.
		// If JSON is valid, we still return the output.
		var parseJSON []interface{}
		if jsonErr := json.Unmarshal(output, &parseJSON); jsonErr == nil {
			return ok(string(output))
		}
		return err(fmt.Sprintf("ast-grep match failed: %v\nOutput: %s", runErr, string(output)))
	}

	return ok(string(output))
}

// HandleFindCode performs structural code searches using an ast-grep pattern.
func HandleFindCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ := getString(args, "pattern")
	path, _ := getString(args, "path", "dir_path", "directory")
	language, _ := getString(args, "language", "lang")

	if pattern == "" {
		return err("pattern parameter is required")
	}
	if path == "" {
		path = "."
	}

	if _, errPath := exec.LookPath("sg"); errPath != nil {
		return err("ast-grep command 'sg' not found in PATH.")
	}

	cmdArgs := []string{"run", "--pattern", pattern}
	if language != "" {
		cmdArgs = append(cmdArgs, "--lang", language)
	}
	cmdArgs = append(cmdArgs, path, "--json")

	// Timeout protection
	tCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(tCtx, "sg", cmdArgs...)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("ast-grep find failed: %v\nOutput: %s", runErr, string(output)))
	}

	return ok(string(output))
}

// HandleFindCodeByRule performs structural searches using a full YAML rule definition.
func HandleFindCodeByRule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ruleYaml, _ := getString(args, "rule", "rule_yaml", "yaml")
	path, _ := getString(args, "path", "dir_path", "directory")

	if ruleYaml == "" {
		return err("rule parameter (YAML rule string) is required")
	}
	if path == "" {
		path = "."
	}

	if _, errPath := exec.LookPath("sg"); errPath != nil {
		return err("ast-grep command 'sg' not found in PATH.")
	}

	// Write rule to a temp file
	tmpRule, errTmp := os.CreateTemp("", "ast-grep-rule-*.yaml")
	if errTmp != nil {
		return err(fmt.Sprintf("failed to create temporary rule file: %v", errTmp))
	}
	rulePath := tmpRule.Name()
	defer os.Remove(rulePath)

	if _, errWrite := tmpRule.Write([]byte(ruleYaml)); errWrite != nil {
		tmpRule.Close()
		return err(fmt.Sprintf("failed to write temporary rule file: %v", errWrite))
	}
	tmpRule.Close()

	// Timeout protection
	tCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(tCtx, "sg", "scan", "--rule", rulePath, path, "--json")
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		// Check if it's just normal matching exit status or an actual failure
		var parseJSON []interface{}
		if jsonErr := json.Unmarshal(output, &parseJSON); jsonErr == nil {
			return ok(string(output))
		}
		return err(fmt.Sprintf("ast-grep scan failed: %v\nOutput: %s", runErr, string(output)))
	}

	return ok(string(output))
}
