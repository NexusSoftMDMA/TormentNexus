package safety

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassifyPath(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		projectDir string
		want       FileScope
	}{
		{"file in project", "/home/user/project/.mcp.json", "/home/user/project", ScopeProject},
		{"file outside project", "/home/user/.config/arc-sync/config.json", "/home/user/project", ScopeUser},
		{"nested file in project", "/home/user/project/sub/dir/file", "/home/user/project", ScopeProject},
		{"sibling directory", "/home/user/other-project/file", "/home/user/project", ScopeUser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyPath(tt.filePath, tt.projectDir)
			if got != tt.want {
				t.Errorf("ClassifyPath(%q, %q) = %q, want %q", tt.filePath, tt.projectDir, got, tt.want)
			}
		})
	}
}

func TestCheckGitignoreNotGitRepo(t *testing.T) {
	dir := t.TempDir()

	warnings := CheckGitignore(dir, ".mcp.json")
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Level != "info" {
		t.Errorf("expected info level, got %q", warnings[0].Level)
	}
	if !strings.Contains(warnings[0].Message, "Not a git repo") {
		t.Errorf("unexpected message: %s", warnings[0].Message)
	}
}

func TestCheckGitignoreNoGitignoreFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	warnings := CheckGitignore(dir, ".mcp.json")
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Level != "warn" {
		t.Errorf("expected warn level, got %q", warnings[0].Level)
	}
	if !strings.Contains(warnings[0].Message, "No .gitignore") {
		t.Errorf("unexpected message: %s", warnings[0].Message)
	}
	if warnings[0].Fix == "" {
		t.Error("expected a fix suggestion")
	}
}

func TestCheckGitignoreFileNotIgnored(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules\n*.log\n"), 0644); err != nil {
		t.Fatal(err)
	}

	warnings := CheckGitignore(dir, ".mcp.json")
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Level != "warn" {
		t.Errorf("expected warn level, got %q", warnings[0].Level)
	}
	if !strings.Contains(warnings[0].Message, "NOT in .gitignore") {
		t.Errorf("unexpected message: %s", warnings[0].Message)
	}
}

func TestCheckGitignoreFileIsIgnored(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules\n.mcp.json\n*.log\n"), 0644); err != nil {
		t.Fatal(err)
	}

	warnings := CheckGitignore(dir, ".mcp.json")
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Level != "info" {
		t.Errorf("expected info level, got %q", warnings[0].Level)
	}
	if !strings.Contains(warnings[0].Message, "gitignored") {
		t.Errorf("unexpected message: %s", warnings[0].Message)
	}
}

func TestCheckGitignoreWithLeadingSlash(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("/.mcp.json\n"), 0644); err != nil {
		t.Fatal(err)
	}

	warnings := CheckGitignore(dir, ".mcp.json")
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Level != "info" {
		t.Errorf("expected info level for /.mcp.json pattern, got %q", warnings[0].Level)
	}
}

func TestCheckGitignoreWithComments(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("# MCP config\n.mcp.json\n"), 0644); err != nil {
		t.Fatal(err)
	}

	warnings := CheckGitignore(dir, ".mcp.json")
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Level != "info" {
		t.Errorf("expected info level, got %q", warnings[0].Level)
	}
}

func TestFormatChangeSummary(t *testing.T) {
	changes := []PlannedChange{
		{Path: "/home/user/project/.mcp.json", Description: "adding 2 servers", Scope: ScopeProject},
		{Path: "/home/user/.config/arc-sync/state.json", Description: "updating skip list", Scope: ScopeUser},
	}

	output := FormatChangeSummary(changes, "/home/user/project")
	if !strings.Contains(output, "PROJECT FILES") {
		t.Error("expected PROJECT FILES section")
	}
	if !strings.Contains(output, "USER FILES") {
		t.Error("expected USER FILES section")
	}
	if !strings.Contains(output, ".mcp.json") {
		t.Error("expected .mcp.json in output")
	}
	if !strings.Contains(output, "state.json") {
		t.Error("expected state.json in output")
	}
}

func TestFormatWarnings(t *testing.T) {
	warnings := []Warning{
		{Level: "warn", Message: "file is not gitignored", Fix: "echo '.mcp.json' >> .gitignore"},
		{Level: "info", Message: "config dir is outside project"},
	}

	output := FormatWarnings(warnings)
	if !strings.Contains(output, "⚠") {
		t.Error("expected warning symbol in output")
	}
	if !strings.Contains(output, "✓") {
		t.Error("expected info symbol in output")
	}
	if !strings.Contains(output, "Fix:") {
		t.Error("expected fix suggestion in output")
	}
}
