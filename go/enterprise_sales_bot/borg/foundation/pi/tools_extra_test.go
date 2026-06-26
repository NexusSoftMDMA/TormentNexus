package pi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGrepToolFindsPattern(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	mustExec(t, runtime, "write", WriteToolInput{Path: "a.txt", Content: "hello world\nhello pi\ngoodbye world\n"})
	result := mustExecResult(t, runtime, "grep", GrepToolInput{Pattern: "hello"})

	text := textFromResult(t, result)
	if !strings.Contains(text, "a.txt") {
		t.Fatalf("expected file path in grep output, got %q", text)
	}
	if !strings.Contains(text, "hello") {
		t.Fatalf("expected pattern in grep output, got %q", text)
	}
}

func TestGrepToolContextLines(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	content := ""
	for i := 1; i <= 5; i++ {
		content += "line " + string(rune('0'+i)) + "\n"
	}
	mustExec(t, runtime, "write", WriteToolInput{Path: "b.txt", Content: content})

	result := mustExecResult(t, runtime, "grep", GrepToolInput{Pattern: `line 3`, Literal: true, Context: 1})
	text := textFromResult(t, result)
	if !strings.Contains(text, "line 3") {
		t.Fatalf("expected match in context output, got %q", text)
	}
}

func TestGrepToolNoMatches(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	mustExec(t, runtime, "write", WriteToolInput{Path: "c.txt", Content: "nothing here\n"})
	result := mustExecResult(t, runtime, "grep", GrepToolInput{Pattern: "zxcvzxcv"})
	text := textFromResult(t, result)
	if text != "No matches found" {
		t.Fatalf("expected no matches message, got %q", text)
	}
}

func TestFindToolFindsFiles(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "util.go"), []byte("package src"), 0o644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0o644)

	result := mustExecResult(t, runtime, "find", FindToolInput{Pattern: "*.go"})
	text := textFromResult(t, result)
	if !strings.Contains(text, ".go") {
		t.Fatalf("expected .go files, got %q", text)
	}
}

func TestFindToolNoFiles(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi"), 0o644)
	result := mustExecResult(t, runtime, "find", FindToolInput{Pattern: "*.go"})
	text := textFromResult(t, result)
	if text != "No files found matching pattern" {
		t.Fatalf("expected no files message, got %q", text)
	}
}

func TestLsToolListsDirectory(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("b"), 0o644)

	result := mustExecResult(t, runtime, "ls", LsToolInput{})
	text := textFromResult(t, result)

	if !strings.Contains(text, "a.txt") {
		t.Fatalf("expected a.txt in ls output, got %q", text)
	}
	if !strings.Contains(text, "b.go") {
		t.Fatalf("expected b.go in ls output, got %q", text)
	}
	if !strings.Contains(text, "subdir/") {
		t.Fatalf("expected subdir/ with slash suffix, got %q", text)
	}
}

func TestLsToolEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	subdir := filepath.Join(dir, "empty")
	os.MkdirAll(subdir, 0o755)

	result := mustExecResult(t, runtime, "ls", LsToolInput{Path: "empty"})
	text := textFromResult(t, result)
	if text != "(empty directory)" {
		t.Fatalf("expected empty directory message, got %q", text)
	}
}

func TestLsToolNotFound(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	result := mustExecResult(t, runtime, "ls", LsToolInput{Path: "nonexistent"})
	text := textFromResult(t, result)
	if !strings.Contains(text, "path not found") {
		t.Fatalf("expected path not found error, got %q", text)
	}
}

func TestGrepToolWithGlob(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "a.go"), []byte("hello world"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "b.txt"), []byte("hello skip me"), 0o644)

	result := mustExecResult(t, runtime, "grep", GrepToolInput{Pattern: "hello", Glob: "*.go"})
	text := textFromResult(t, result)
	if strings.Contains(text, "b.txt") {
		t.Fatalf("expected only .go files, got %q", text)
	}
	if !strings.Contains(text, "a.go") {
		t.Fatalf("expected a.go match, got %q", text)
	}
}

func TestFindToolWithSubdir(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main2"), 0o644)

	result := mustExecResult(t, runtime, "find", FindToolInput{Pattern: "*.go", Path: "src"})
	text := textFromResult(t, result)
	if text == "No files found matching pattern" {
		t.Fatal("expected to find at least one .go file under src")
	}
}

func TestLsToolLimit(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	for i := 0; i < 10; i++ {
		fname := "file" + string(rune('0'+i)) + ".txt"
		os.WriteFile(filepath.Join(dir, fname), []byte("x"), 0o644)
	}

	result := mustExecResult(t, runtime, "ls", LsToolInput{Limit: 3})
	details, ok := result.Details.(*LsToolDetails)
	if !ok || details == nil {
		t.Fatalf("expected ls details with limit info, got %#v", result.Details)
	}
	if details.EntryLimitReached != 3 {
		t.Fatalf("expected entry limit 3, got %v", details.EntryLimitReached)
	}
}

func TestGrepToolLimit(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	content := ""
	for i := 0; i < 10; i++ {
		content += "line match\n"
	}
	mustExec(t, runtime, "write", WriteToolInput{Path: "many.txt", Content: content})

	result := mustExecResult(t, runtime, "grep", GrepToolInput{Pattern: "match", Limit: 3})
	details, ok := result.Details.(*GrepToolDetails)
	if !ok || details == nil {
		t.Fatalf("expected grep details with limit info, got %#v", result.Details)
	}
	if details.MatchLimitReached != 3 {
		t.Fatalf("expected match limit 3, got %v", details.MatchLimitReached)
	}
}

func TestFindToolResultLimit(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	for i := 0; i < 10; i++ {
		fname := "f" + string(rune('0'+i)) + ".txt"
		os.WriteFile(filepath.Join(dir, fname), []byte("x"), 0o644)
	}

	result := mustExecResult(t, runtime, "find", FindToolInput{Pattern: "*.txt", Limit: 3})
	details, ok := result.Details.(*FindToolDetails)
	if !ok || details == nil {
		t.Fatalf("expected find details with limit info, got %#v", result.Details)
	}
	if details.ResultLimitReached != 3 {
		t.Fatalf("expected result limit 3, got %v", details.ResultLimitReached)
	}
}

// Helpers reused from runtime_test.go: mustExec, mustExecResult, textFromResult
