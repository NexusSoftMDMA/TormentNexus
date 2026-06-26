package pi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestReadToolReportsContinuationDetails(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))
	contentLines := make([]string, 0, 2105)
	for i := 1; i <= 2105; i++ {
		contentLines = append(contentLines, fmt.Sprintf("line-%04d", i))
	}
	writeInput, _ := json.Marshal(WriteToolInput{Path: "large.txt", Content: strings.Join(contentLines, "\n")})
	if _, err := runtime.ExecuteTool(context.Background(), "", "write", writeInput, nil); err != nil {
		t.Fatal(err)
	}
	readInput, _ := json.Marshal(ReadToolInput{Path: "large.txt"})
	result, err := runtime.ExecuteTool(context.Background(), "", "read", readInput, nil)
	if err != nil {
		t.Fatal(err)
	}
	text := textFromResult(t, result)
	if !strings.Contains(text, "Use offset=2001 to continue") {
		t.Fatalf("expected continuation hint, got %q", text)
	}
	details, ok := result.Details.(*ReadToolDetails)
	if !ok || details == nil || details.Truncation == nil {
		t.Fatalf("expected read truncation details, got %#v", result.Details)
	}
	if !details.Truncation.Truncated || details.Truncation.ContinuationOffset != 2001 {
		t.Fatalf("unexpected truncation details: %#v", details.Truncation)
	}
}

func TestBashToolReturnsErrorPayloadWithExitCode(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))
	input, _ := json.Marshal(BashToolInput{Command: failingShellCommand()})
	result, err := runtime.ExecuteTool(context.Background(), "", "bash", input, nil)
	if err == nil {
		t.Fatal("expected bash error")
	}
	if result == nil || !result.IsError {
		t.Fatalf("expected error result, got %#v", result)
	}
	text := textFromResult(t, result)
	if !strings.Contains(text, "Command exited with code") {
		t.Fatalf("expected exit code message, got %q", text)
	}
}

func failingShellCommand() string {
	if isWindows() {
		return "exit /b 7"
	}
	return "exit 7"
}
