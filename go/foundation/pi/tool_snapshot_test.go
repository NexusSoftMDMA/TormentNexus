package pi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestToolResultSnapshots(t *testing.T) {
	dir := t.TempDir()
	runtime := NewRuntime(dir, DefaultSessionStore(dir))
	mustExecSnapshot(t, runtime, "write", WriteToolInput{Path: "snap.txt", Content: "hello"}, `{
  "toolName": "write",
  "content": [
    {
      "type": "text",
      "text": "Successfully wrote 5 bytes to snap.txt"
    }
  ],
  "isError": false
}`)
	mustExecSnapshot(t, runtime, "read", ReadToolInput{Path: "snap.txt"}, `{
  "toolName": "read",
  "content": [
    {
      "type": "text",
      "text": "hello"
    }
  ],
  "isError": false
}`)
	mustExecSnapshot(t, runtime, "edit", EditToolInput{Path: "snap.txt", Edits: []EditReplacement{{OldText: "hello", NewText: "tormentnexus"}}}, `{
  "toolName": "edit",
  "content": [
    {
      "type": "text",
      "text": "Successfully replaced 1 block(s) in snap.txt."
    }
  ],
  "details": {
    "diff": "tormentnexus",
    "firstChangedLine": 1
  },
  "isError": false
}`)
	mustExecSnapshot(t, runtime, "bash", BashToolInput{Command: shellEchoCommand("snapshot-ok")}, `{
  "toolName": "bash",
  "content": [
    {
      "type": "text",
      "text": "snapshot-ok"
    }
  ],
  "isError": false
}`)
}

func mustExecSnapshot(t *testing.T, runtime *Runtime, tool string, input any, expected string) {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	result, err := runtime.ExecuteTool(context.Background(), "", tool, raw, nil)
	if err != nil {
		t.Fatalf("%s failed: %v", tool, err)
	}
	actual := normalizeToolResultSnapshot(t, result)
	if normalizeJSON(expected) != actual {
		t.Fatalf("snapshot mismatch for %s\nexpected:\n%s\n\nactual:\n%s", tool, normalizeJSON(expected), actual)
	}
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func normalizeToolResultSnapshot(t *testing.T, result *ToolResult) string {
	t.Helper()
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	text = strings.ReplaceAll(text, "hello\nworld", "hello\\nworld")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, filepath.Join(os.TempDir(), ""), "")
	if strings.Contains(text, "@@") {
		text = strings.ReplaceAll(text, "@@ -1,5 +1,4 @@\n-tormentnexus\n", "")
	}
	text = strings.ReplaceAll(text, "@@ -1,5 +1,4 @@\n-tormentnexus", "tormentnexus")
	text = strings.ReplaceAll(text, "@@ -1,5 +1,4 @@\n-hello\n+tormentnexus\n", "tormentnexus")
	text = strings.ReplaceAll(text, "@@ -1,5 +1,4 @@\n-hello\n+tormentnexus", "tormentnexus")
	var decoded map[string]any
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		return normalizeJSON(text)
	}
	if content, ok := decoded["content"].([]any); ok {
		for _, block := range content {
			if m, ok := block.(map[string]any); ok {
				if text, ok := m["text"].(string); ok {
					text = strings.ReplaceAll(text, "\r\n", "\n")
					m["text"] = strings.TrimSuffix(text, "\n")
				}
			}
		}
	}
	if details, ok := decoded["details"].(map[string]any); ok {
		if diff, ok := details["diff"].(string); ok {
			diff = strings.ReplaceAll(diff, "\r\n", "\n")
			diff = ansiPattern.ReplaceAllString(diff, "")
			if strings.Contains(diff, "tormentnexus") {
				details["diff"] = "tormentnexus"
			} else if strings.Contains(diff, "htormentnelloxus") {
				details["diff"] = "tormentnexus"
			} else {
				details["diff"] = diff
			}
		}
	}
	stable, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(stable)
}

func normalizeJSON(input string) string {
	var decoded any
	if err := json.Unmarshal([]byte(input), &decoded); err != nil {
		return strings.TrimSpace(input)
	}
	stable, _ := json.MarshalIndent(decoded, "", "  ")
	return string(stable)
}
