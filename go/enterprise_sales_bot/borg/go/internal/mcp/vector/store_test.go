package vector

import (
	"encoding/json"
	"math"
	"testing"
)

func openTestStore(t *testing.T) *VectorStore {
	t.Helper()
	vs, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { vs.Close() })
	return vs
}

func mustTool(id, server, name, desc, category, tags string) ToolRecord {
	return ToolRecord{ID: id, ServerName: server, ToolName: name, Description: desc, Category: category, Tags: tags, SchemaJSON: `{"type":"object","properties":{}}`, Source: "test", Version: "1.0.0"}
}

func mustEmbed(t *testing.T, vs *VectorStore, toolID string, vec []float32) {
	t.Helper()
	if vs.StoreEmbedding(EmbeddingRecord{ToolID: toolID, ModelName: "all-MiniLM-L6-v2", Dimension: len(vec), Vector: vec}) != nil {
		t.Fatalf("embed")
	}
}

func TestUpsertAndGetTool(t *testing.T) {
	vs := openTestStore(t)
	tool := mustTool("srv::read_file", "filesystem", "read_file", "Read file contents", "filesystem", "file,io")
	if vs.UpsertTool(tool) != nil {
		t.Fatal("upsert")
	}
	got, _ := vs.GetTool("srv::read_file")
	if got == nil {
		t.Fatal("nil")
	}
	if got.ToolName != "read_file" {
		t.Errorf("got %q", got.ToolName)
	}
}

func TestDeleteTool(t *testing.T) {
	vs := openTestStore(t)
	vs.UpsertTool(mustTool("srv::del", "fs", "del", "delete", "fs", ""))
	vs.DeleteTool("srv::del")
	got, _ := vs.GetTool("srv::del")
	if got != nil {
		t.Error("expected nil")
	}
}

func TestStoreAndGetEmbedding(t *testing.T) {
	vs := openTestStore(t)
	vs.UpsertTool(mustTool("srv::e", "fs", "e", "test", "fs", ""))
	orig := []float32{0.1, 0.2, 0.3, 0.4}
	mustEmbed(t, vs, "srv::e", orig)
	got, _ := vs.GetEmbedding("srv::e", "all-MiniLM-L6-v2")
	if got == nil {
		t.Fatal("nil")
	}
	for i, v := range got {
		if math.Abs(float64(v-orig[i])) > 1e-6 {
			t.Errorf("vec[%d]: %f != %f", i, v, orig[i])
		}
	}
}

func TestSemanticSearch(t *testing.T) {
	vs := openTestStore(t)
	for _, tool := range []ToolRecord{
		mustTool("fs::read", "fs", "read_file", "Read a file from disk", "filesystem", "file"),
		mustTool("fs::write", "fs", "write_file", "Write content to a file", "filesystem", "file"),
		mustTool("web::fetch", "web", "fetch_url", "Fetch content from a URL", "browser", "http"),
		mustTool("db::query", "db", "query_sql", "Execute a SQL query", "database", "sql"),
	} {
		vs.UpsertTool(tool)
	}
	mustEmbed(t, vs, "fs::read", []float32{0.9, 0.1, 0, 0})
	mustEmbed(t, vs, "fs::write", []float32{0.8, 0.2, 0, 0})
	mustEmbed(t, vs, "web::fetch", []float32{0.1, 0, 0.9, 0})
	mustEmbed(t, vs, "db::query", []float32{0, 0.1, 0, 0.9})
	results, err := vs.Search(SearchQuery{QueryVec: []float32{1, 0, 0, 0}, TopK: 3, MinScore: 0.1})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected 2+, got %d", len(results))
	}
	if results[0].Tool.ToolName != "read_file" {
		t.Errorf("top: %q", results[0].Tool.ToolName)
	}
}

func TestCategoryFilter(t *testing.T) {
	vs := openTestStore(t)
	vs.UpsertTool(mustTool("fs::cat", "fs", "cat", "cat file", "filesystem", ""))
	vs.UpsertTool(mustTool("db::sel", "db", "select", "select data", "database", ""))
	mustEmbed(t, vs, "fs::cat", []float32{1, 0})
	mustEmbed(t, vs, "db::sel", []float32{0.9, 0.1})
	results, _ := vs.Search(SearchQuery{QueryVec: []float32{1, 0}, TopK: 10, MinScore: 0, Categories: []string{"database"}})
	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}
	if results[0].Tool.ToolName != "select" {
		t.Errorf("got %q", results[0].Tool.ToolName)
	}
}

func TestKeywordFallback(t *testing.T) {
	vs := openTestStore(t)
	vs.UpsertTool(mustTool("fs::grep", "fs", "grep", "Search files for patterns", "filesystem", ""))
	vs.UpsertTool(mustTool("fs::ls", "fs", "ls", "List directory", "filesystem", ""))
	results, _ := vs.Search(SearchQuery{QueryText: "files", TopK: 5})
	if len(results) == 0 {
		t.Fatal("no results")
	}
	if results[0].Tool.ToolName != "grep" {
		t.Errorf("got %q", results[0].Tool.ToolName)
	}
}

func TestRecordUsage(t *testing.T) {
	vs := openTestStore(t)
	vs.UpsertTool(mustTool("fs::u", "fs", "u", "test", "fs", ""))
	vs.RecordUsage("fs::u", true)
	vs.RecordUsage("fs::u", true)
	mustEmbed(t, vs, "fs::u", []float32{0.5, 0.5})
	results, _ := vs.Search(SearchQuery{QueryVec: []float32{0.5, 0.5}, TopK: 5, MinScore: 0})
	if len(results) == 0 {
		t.Fatal("no results")
	}
	if !results[0].Boosted {
		t.Error("expected boosted")
	}
}

func TestCosineSim(t *testing.T) {
	tests := []struct {
		a, b []float32
		e    float64
	}{
		{[]float32{1, 0}, []float32{1, 0}, 1.0},
		{[]float32{1, 0}, []float32{0, 1}, 0.0},
		{[]float32{}, []float32{}, 0.0},
	}
	for _, tt := range tests {
		if math.Abs(cosineSim(tt.a, tt.b)-tt.e) > 1e-9 {
			t.Errorf("cos(%v,%v)=%f != %f", tt.a, tt.b, cosineSim(tt.a, tt.b), tt.e)
		}
	}
}

func TestToolSchemaJSON(t *testing.T) {
	vs := openTestStore(t)
	schema := map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}}
	sj, _ := json.Marshal(schema)
	tool := ToolRecord{ID: "fs::j", ServerName: "fs", ToolName: "j", SchemaJSON: string(sj), Source: "test"}
	vs.UpsertTool(tool)
	got, _ := vs.GetTool("fs::j")
	var parsed map[string]interface{}
	if json.Unmarshal([]byte(got.SchemaJSON), &parsed) != nil {
		t.Fatal("json")
	}
	if parsed["type"] != "object" {
		t.Error("type")
	}
}
