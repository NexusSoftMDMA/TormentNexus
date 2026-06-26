package repomap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateBuildsRankedRepoMap(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n\nfunc helper() {}\n")
	mustWrite(t, filepath.Join(dir, "pkg", "worker.go"), "package pkg\n\ntype Worker struct {}\n\nfunc Run() {}\n")
	mustWrite(t, filepath.Join(dir, "pkg", "worker_test.go"), "package pkg\n\nfunc TestWorker() {}\n")

	result, err := Generate(Options{BaseDir: dir, MentionedIdents: []string{"Worker"}, MaxFiles: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 non-test source entries, got %d", len(result.Entries))
	}
	if result.Entries[0].Path != "pkg/worker.go" {
		t.Fatalf("expected worker.go to rank first, got %s", result.Entries[0].Path)
	}
	if !strings.Contains(result.Map, "pkg/worker.go") || !strings.Contains(result.Map, "struct Worker") || !strings.Contains(result.Map, "func Run") {
		t.Fatalf("unexpected repo map output:\n%s", result.Map)
	}
	if !strings.HasPrefix(result.Map, "<repo_map>\n") || !strings.HasSuffix(result.Map, "</repo_map>") {
		t.Fatalf("repo map missing markers: %s", result.Map)
	}
}

func TestGenerateUsesReferenceGraphGroundworkForRanking(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {\n  _ = Worker{}\n  Run()\n}\n")
	mustWrite(t, filepath.Join(dir, "worker.go"), "package main\n\ntype Worker struct {}\n\nfunc Run() {}\n")
	result, err := Generate(Options{BaseDir: dir, MaxFiles: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 source entries, got %d", len(result.Entries))
	}
	if result.Entries[0].Path != "worker.go" {
		t.Fatalf("expected worker.go to rank first from reference graph, got %#v", result.Entries)
	}
}

func TestGenerateIncludesTestsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "main_test.go"), "package main\n\nfunc TestMain() {}\n")
	result, err := Generate(Options{BaseDir: dir, IncludeTests: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 1 || result.Entries[0].Path != "main_test.go" {
		t.Fatalf("unexpected entries: %#v", result.Entries)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
