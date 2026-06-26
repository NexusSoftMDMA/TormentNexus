package pi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeExecuteReadWriteEditAndBashTools(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))

	writeInput, _ := json.Marshal(WriteToolInput{Path: "notes.txt", Content: "hello\nworld\n"})
	writeResult, err := runtime.ExecuteTool(context.Background(), "", "write", writeInput, nil)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if got := textFromResult(t, writeResult); !strings.Contains(got, "Successfully wrote") {
		t.Fatalf("unexpected write result: %s", got)
	}

	readInput, _ := json.Marshal(ReadToolInput{Path: "notes.txt"})
	readResult, err := runtime.ExecuteTool(context.Background(), "", "read", readInput, nil)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if got := textFromResult(t, readResult); got != "hello\nworld\n" && got != "hello\nworld" {
		t.Fatalf("unexpected read result: %q", got)
	}

	editInput, _ := json.Marshal(EditToolInput{Path: "notes.txt", Edits: []EditReplacement{{OldText: "world", NewText: "tormentnexus"}}})
	editResult, err := runtime.ExecuteTool(context.Background(), "", "edit", editInput, nil)
	if err != nil {
		t.Fatalf("edit failed: %v", err)
	}
	if got := textFromResult(t, editResult); !strings.Contains(got, "Successfully replaced 1 block(s) in notes.txt.") {
		t.Fatalf("unexpected edit result: %s", got)
	}
	data, err := os.ReadFile(filepath.Join(dir, "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "tormentnexus") {
		t.Fatalf("expected file to contain tormentnexus, got %q", string(data))
	}

	bashInput, _ := json.Marshal(BashToolInput{Command: shellEchoCommand("foundation-ok")})
	bashResult, err := runtime.ExecuteTool(context.Background(), "", "bash", bashInput, nil)
	if err != nil {
		t.Fatalf("bash failed: %v", err)
	}
	if got := textFromResult(t, bashResult); !strings.Contains(got, "foundation-ok") {
		t.Fatalf("unexpected bash result: %q", got)
	}
}

func TestRuntimePersistsSessionToolRuns(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))
	session, err := runtime.CreateSession("test-session")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.AppendUserText(session.Metadata.SessionID, "inspect go.mod"); err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(WriteToolInput{Path: "session.txt", Content: "persisted"})
	if _, err := runtime.ExecuteTool(context.Background(), session.Metadata.SessionID, "write", input, nil); err != nil {
		t.Fatal(err)
	}
	loaded, err := runtime.LoadSession(session.Metadata.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 3 {
		t.Fatalf("expected 3 session entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[1].Kind != "tool_call" || loaded.Entries[2].Kind != "tool_result" {
		t.Fatalf("unexpected session entries: %#v", loaded.Entries)
	}
}

func TestRuntimeEmitsOrderedEvents(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))
	input, _ := json.Marshal(WriteToolInput{Path: "ordered.txt", Content: "x"})
	var seen []RunEventType
	_, err := runtime.ExecuteTool(context.Background(), "", "write", input, func(event RunEvent) {
		seen = append(seen, event.Type)
	})
	if err != nil {
		t.Fatal(err)
	}
	wanted := []RunEventType{EventAgentStart, EventTurnStart, EventMessageStart, EventMessageEnd, EventToolExecutionStart, EventToolExecutionEnd, EventTurnEnd, EventAgentEnd}
	if len(seen) != len(wanted) {
		t.Fatalf("unexpected event count: got %v want %v", seen, wanted)
	}
	for i := range wanted {
		if seen[i] != wanted[i] {
			t.Fatalf("event %d mismatch: got %s want %s", i, seen[i], wanted[i])
		}
	}
}

func textFromResult(t *testing.T, result *ToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	block, ok := result.Content[0].(TextContent)
	if ok {
		return block.Text
	}
	mapBlock, ok := result.Content[0].(map[string]any)
	if ok {
		if text, _ := mapBlock["text"].(string); text != "" {
			return text
		}
	}
	t.Fatalf("unexpected content block type: %#v", result.Content[0])
	return ""
}

func shellEchoCommand(text string) string {
	if isWindows() {
		return "echo " + text
	}
	return "printf '%s' " + text
}

func mustExec(t *testing.T, runtime *Runtime, tool string, input any) {
	raw, _ := json.Marshal(input)
	if _, err := runtime.ExecuteTool(context.Background(), "", tool, raw, nil); err != nil {
		t.Fatalf("unexpected error running %s: %v", tool, err)
	}
}

func mustExecResult(t *testing.T, runtime *Runtime, tool string, input any) *ToolResult {
	raw, _ := json.Marshal(input)
	result, err := runtime.ExecuteTool(context.Background(), "", tool, raw, nil)
	if err != nil {
		t.Fatalf("unexpected error running %s: %v", tool, err)
	}
	if result == nil {
		t.Fatalf("expected result from %s, got nil", tool)
	}
	return result
}
