package mcp

import (
	"encoding/json"
	"testing"
)

func TestPruneSchema_RemovesMetadataKeys(t *testing.T) {
	schema := json.RawMessage(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"$id": "test",
		"title": "Test Schema",
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"title": "Name Field",
				"examples": ["foo", "bar"],
				"description": "The name"
			}
		},
		"required": ["name"],
		"x-custom": "vendor"
	}`)

	result := PruneSchema(schema)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(result, &obj); err != nil {
		t.Fatalf("Failed to unmarshal pruned schema: %v", err)
	}

	// Should be removed
	for _, key := range []string{"$schema", "$id", "title", "x-custom"} {
		if _, ok := obj[key]; ok {
			t.Errorf("Expected key %q to be removed", key)
		}
	}

	// Should be kept
	for _, key := range []string{"type", "properties", "required"} {
		if _, ok := obj[key]; !ok {
			t.Errorf("Expected key %q to be preserved", key)
		}
	}

	// Check nested property pruning
	var props map[string]json.RawMessage
	if err := json.Unmarshal(obj["properties"], &props); err != nil {
		t.Fatalf("Failed to unmarshal properties: %v", err)
	}
	var nameProp map[string]json.RawMessage
	if err := json.Unmarshal(props["name"], &nameProp); err != nil {
		t.Fatalf("Failed to unmarshal name property: %v", err)
	}

	if _, ok := nameProp["title"]; ok {
		t.Error("Expected nested 'title' to be removed")
	}
	if _, ok := nameProp["examples"]; ok {
		t.Error("Expected nested 'examples' to be removed")
	}
	if _, ok := nameProp["type"]; !ok {
		t.Error("Expected nested 'type' to be preserved")
	}
	if _, ok := nameProp["description"]; !ok {
		t.Error("Expected nested 'description' to be preserved")
	}
}

func TestPruneSchema_PreservesValidationKeywords(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"count": {
				"type": "integer",
				"minimum": 1,
				"maximum": 100,
				"default": 10,
				"deprecated": true
			},
			"status": {
				"type": "string",
				"enum": ["active", "inactive"],
				"const": "active",
				"pattern": "^[a-z]+$",
				"format": "email"
			}
		},
		"required": ["count"],
		"additionalProperties": false
	}`)

	result := PruneSchema(schema)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(result, &obj); err != nil {
		t.Fatalf("Failed to unmarshal pruned schema: %v", err)
	}

	// Top-level validation keywords preserved
	for _, key := range []string{"type", "properties", "required", "additionalProperties"} {
		if _, ok := obj[key]; !ok {
			t.Errorf("Expected key %q to be preserved", key)
		}
	}

	// Check property-level validation
	var props map[string]json.RawMessage
	if err := json.Unmarshal(obj["properties"], &props); err != nil {
		t.Fatalf("Failed to unmarshal properties: %v", err)
	}

	var count map[string]json.RawMessage
	if err := json.Unmarshal(props["count"], &count); err != nil {
		t.Fatalf("Failed to unmarshal count property: %v", err)
	}
	for _, key := range []string{"type", "minimum", "maximum", "default"} {
		if _, ok := count[key]; !ok {
			t.Errorf("count: expected key %q to be preserved", key)
		}
	}
	if _, ok := count["deprecated"]; ok {
		t.Error("count: expected 'deprecated' to be removed")
	}

	var status map[string]json.RawMessage
	if err := json.Unmarshal(props["status"], &status); err != nil {
		t.Fatalf("Failed to unmarshal status property: %v", err)
	}
	for _, key := range []string{"type", "enum", "const", "pattern", "format"} {
		if _, ok := status[key]; !ok {
			t.Errorf("status: expected key %q to be preserved", key)
		}
	}
}

func TestPruneSchema_HandlesAnyOf(t *testing.T) {
	schema := json.RawMessage(`{
		"anyOf": [
			{"type": "string", "title": "String Option"},
			{"type": "null", "title": "Null Option"}
		]
	}`)

	result := PruneSchema(schema)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(result, &obj); err != nil {
		t.Fatalf("Failed to unmarshal pruned schema: %v", err)
	}

	var anyOf []json.RawMessage
	if err := json.Unmarshal(obj["anyOf"], &anyOf); err != nil {
		t.Fatalf("Failed to unmarshal anyOf: %v", err)
	}
	if len(anyOf) != 2 {
		t.Fatalf("Expected 2 anyOf items, got %d", len(anyOf))
	}

	// Each item should have title removed
	for i, item := range anyOf {
		var itemObj map[string]json.RawMessage
		if err := json.Unmarshal(item, &itemObj); err != nil {
			t.Fatalf("anyOf[%d]: failed to unmarshal: %v", i, err)
		}
		if _, ok := itemObj["title"]; ok {
			t.Errorf("anyOf[%d]: expected 'title' to be removed", i)
		}
		if _, ok := itemObj["type"]; !ok {
			t.Errorf("anyOf[%d]: expected 'type' to be preserved", i)
		}
	}
}

func TestPruneSchema_NonObjectPassthrough(t *testing.T) {
	// Non-object schemas should pass through unchanged
	cases := []string{
		`"string"`,
		`42`,
		`true`,
		`null`,
		`[1,2,3]`,
	}
	for _, tc := range cases {
		result := PruneSchema(json.RawMessage(tc))
		if string(result) != tc {
			t.Errorf("Expected %q to pass through, got %q", tc, string(result))
		}
	}
}

func TestPruneSchema_EmptyInput(t *testing.T) {
	result := PruneSchema(nil)
	if result != nil {
		t.Errorf("Expected nil for nil input, got %q", string(result))
	}

	result = PruneSchema(json.RawMessage{})
	if len(result) != 0 {
		t.Errorf("Expected empty for empty input, got %q", string(result))
	}
}

func TestHashTools_Deterministic(t *testing.T) {
	tools := []Tool{
		{Name: "foo", Description: "do foo", InputSchema: json.RawMessage(`{"type":"object"}`)},
		{Name: "bar", Description: "do bar", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}

	hash1 := HashTools(tools)
	hash2 := HashTools(tools)
	if hash1 != hash2 {
		t.Errorf("HashTools not deterministic: %s != %s", hash1, hash2)
	}
	if len(hash1) != 64 {
		t.Errorf("Expected 64-char hex hash, got %d chars", len(hash1))
	}
}

func TestHashTools_ChangesOnModification(t *testing.T) {
	tools1 := []Tool{
		{Name: "foo", Description: "do foo", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}
	tools2 := []Tool{
		{Name: "foo", Description: "do foo updated", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}

	hash1 := HashTools(tools1)
	hash2 := HashTools(tools2)
	if hash1 == hash2 {
		t.Error("Expected different hashes for different tools")
	}
}

func TestAuditTools_CalculatesSizes(t *testing.T) {
	tools := []Tool{
		{Name: "a", Description: "short", InputSchema: json.RawMessage(`{"type":"object"}`)},
		{Name: "b", Description: "a longer description here", InputSchema: json.RawMessage(`{"type":"object","properties":{"x":{"type":"string"}}}`)},
	}

	stats, total := AuditTools(tools)
	if len(stats) != 2 {
		t.Fatalf("Expected 2 stats, got %d", len(stats))
	}

	if stats[0].Name != "a" {
		t.Errorf("Expected first tool name 'a', got %q", stats[0].Name)
	}
	if stats[0].DescChars != 5 {
		t.Errorf("Expected desc chars 5, got %d", stats[0].DescChars)
	}

	expectedTotal := stats[0].TotalChars + stats[1].TotalChars
	if total != expectedTotal {
		t.Errorf("Expected total %d, got %d", expectedTotal, total)
	}
}

func TestExtractJSON_DirectArray(t *testing.T) {
	input := `[{"name":"foo"}]`
	result := extractJSON(input)
	if result != input {
		t.Errorf("Expected %q, got %q", input, result)
	}
}

func TestExtractJSON_MarkdownFenced(t *testing.T) {
	input := "```json\n[{\"name\":\"foo\"}]\n```"
	expected := `[{"name":"foo"}]`
	result := extractJSON(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtractJSON_WithPreamble(t *testing.T) {
	input := "Here are the optimized tools:\n[{\"name\":\"foo\"}]"
	expected := `[{"name":"foo"}]`
	result := extractJSON(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
