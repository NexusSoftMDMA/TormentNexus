package vector

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type VectorStore struct {
	db                *sql.DB
	mu                sync.RWMutex
	stmtUpsertTool    *sql.Stmt
	stmtUpsertEmbed   *sql.Stmt
	stmtGetEmbed      *sql.Stmt
	stmtAllEmbeddings *sql.Stmt
	stmtRecordUsage   *sql.Stmt
}

func Open(path string) (*VectorStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if path != ":memory:" {
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA synchronous=NORMAL")
	}
	if err := InitSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	vs := &VectorStore{db: db}
	if err := vs.prepare(); err != nil {
		db.Close()
		return nil, err
	}
	return vs, nil
}

func (vs *VectorStore) prepare() error {
	var err error
	vs.stmtUpsertTool, err = vs.db.Prepare(`INSERT INTO tools (id, server_name, tool_name, description, schema_json, category, tags, source, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET description=excluded.description, schema_json=excluded.schema_json, category=excluded.category, tags=excluded.tags, version=excluded.version, updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')`)
	if err != nil {
		return err
	}
	vs.stmtUpsertEmbed, err = vs.db.Prepare(`INSERT INTO tool_embeddings (tool_id, model_name, dimension, vector, content_src) VALUES (?, ?, ?, ?, ?) ON CONFLICT(tool_id, model_name) DO UPDATE SET vector=excluded.vector, dimension=excluded.dimension`)
	if err != nil {
		return err
	}
	vs.stmtGetEmbed, err = vs.db.Prepare(`SELECT vector, dimension FROM tool_embeddings WHERE tool_id = ? AND model_name = ?`)
	if err != nil {
		return err
	}
	vs.stmtAllEmbeddings, err = vs.db.Prepare(`SELECT e.tool_id, e.vector, e.dimension, t.server_name, t.tool_name, t.description, t.schema_json, t.category, t.tags, t.source, t.version FROM tool_embeddings e JOIN tools t ON t.id = e.tool_id WHERE e.model_name = ?`)
	if err != nil {
		return err
	}
	vs.stmtRecordUsage, err = vs.db.Prepare(`INSERT INTO tool_usage (tool_id, select_count, success_rate, last_used_at) VALUES (?, 1, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now')) ON CONFLICT(tool_id) DO UPDATE SET select_count = select_count + 1, success_rate = (success_rate * (select_count - 1) + excluded.success_rate) / select_count, last_used_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')`)
	if err != nil {
		return err
	}
	return nil
}

func (vs *VectorStore) Close() error {
	for _, s := range []*sql.Stmt{vs.stmtUpsertTool, vs.stmtUpsertEmbed, vs.stmtGetEmbed, vs.stmtAllEmbeddings, vs.stmtRecordUsage} {
		if s != nil {
			s.Close()
		}
	}
	return vs.db.Close()
}

func (vs *VectorStore) UpsertTool(tool ToolRecord) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	_, err := vs.stmtUpsertTool.Exec(tool.ID, tool.ServerName, tool.ToolName, tool.Description, tool.SchemaJSON, tool.Category, tool.Tags, tool.Source, tool.Version)
	return err
}

func (vs *VectorStore) GetTool(id string) (*ToolRecord, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	var t ToolRecord
	var cAt, uAt string
	err := vs.db.QueryRow(`SELECT id, server_name, tool_name, description, schema_json, category, tags, source, version, created_at, updated_at FROM tools WHERE id = ?`, id).Scan(&t.ID, &t.ServerName, &t.ToolName, &t.Description, &t.SchemaJSON, &t.Category, &t.Tags, &t.Source, &t.Version, &cAt, &uAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, cAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, uAt)
	return &t, nil
}

func (vs *VectorStore) DeleteTool(id string) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	_, err := vs.db.Exec("DELETE FROM tools WHERE id = ?", id)
	return err
}

type EmbeddingRecord struct {
	ToolID    string
	ModelName string
	Dimension int
	Vector    []float32
}

func (vs *VectorStore) StoreEmbedding(rec EmbeddingRecord) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	_, err := vs.stmtUpsertEmbed.Exec(rec.ToolID, rec.ModelName, rec.Dimension, encodeVec(rec.Vector), "description")
	return err
}

func (vs *VectorStore) GetEmbedding(toolID, modelName string) ([]float32, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	var blob []byte
	var dim int
	err := vs.stmtGetEmbed.QueryRow(toolID, modelName).Scan(&blob, &dim)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return decodeVec(blob, dim), nil
}

func (vs *VectorStore) Search(q SearchQuery) ([]SearchResult, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	topK := q.TopK
	if topK <= 0 {
		topK = 10
	}
	minScore := q.MinScore
	if minScore <= 0 {
		minScore = 0.3
	}
	if len(q.QueryVec) > 0 {
		return vs.semanticSearch(q.QueryVec, q.Categories, topK, minScore)
	}
	return vs.keywordSearch(q.QueryText, q.Categories, topK)
}

func (vs *VectorStore) semanticSearch(queryVec []float32, categories []string, topK int, minScore float64) ([]SearchResult, error) {
	rows, err := vs.stmtAllEmbeddings.Query("all-MiniLM-L6-v2")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type scored struct {
		tool    ToolRecord
		score   float64
		boosted bool
	}
	var results []scored
	for rows.Next() {
		var toolID string
		var blob []byte
		var dim int
		var t ToolRecord
		if rows.Scan(&toolID, &blob, &dim, &t.ServerName, &t.ToolName, &t.Description, &t.SchemaJSON, &t.Category, &t.Tags, &t.Source, &t.Version) != nil {
			continue
		}
		t.ID = toolID
		if len(categories) > 0 {
			found := false
			for _, c := range categories {
				if strings.EqualFold(t.Category, c) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		vec := decodeVec(blob, dim)
		sc := cosineSim(queryVec, vec)
		if sc >= minScore {
			results = append(results, scored{tool: t, score: sc})
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].score > results[j].score })
	// Usage boost
	ids := make([]string, 0, len(results))
	for i := range results {
		ids = append(ids, results[i].tool.ID)
	}
	usage := vs.loadUsage(ids)
	for i := range results {
		if u, ok := usage[results[i].tool.ID]; ok && u > 0 {
			results[i].boosted = true
		}
	}
	if len(results) > topK {
		results = results[:topK]
	}
	out := make([]SearchResult, len(results))
	for i, r := range results {
		out[i] = SearchResult{Tool: r.tool, Score: r.score, Rank: i + 1, Boosted: r.boosted}
	}
	return out, nil
}

func (vs *VectorStore) keywordSearch(query string, categories []string, topK int) ([]SearchResult, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	args := []interface{}{pattern, pattern}
	catClause := ""
	if len(categories) > 0 {
		ph := make([]string, len(categories))
		for i, c := range categories {
			ph[i] = "?"
			args = append(args, c)
		}
		catClause = " AND category IN (" + strings.Join(ph, ",") + ")"
	}
	args = append(args, topK)
	rows, err := vs.db.Query("SELECT id, server_name, tool_name, description, schema_json, category, tags, source, version FROM tools WHERE (LOWER(description) LIKE ? OR LOWER(tool_name) LIKE ?)"+catClause+" ORDER BY updated_at DESC LIMIT ?", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SearchResult
	rank := 0
	for rows.Next() {
		rank++
		var t ToolRecord
		if rows.Scan(&t.ID, &t.ServerName, &t.ToolName, &t.Description, &t.SchemaJSON, &t.Category, &t.Tags, &t.Source, &t.Version) != nil {
			continue
		}
		out = append(out, SearchResult{Tool: t, Score: 1.0 / float64(rank), Rank: rank})
	}
	return out, nil
}

func (vs *VectorStore) RecordUsage(toolID string, success bool) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	sr := 0.0
	if success {
		sr = 1.0
	}
	_, err := vs.stmtRecordUsage.Exec(toolID, sr)
	return err
}

func (vs *VectorStore) loadUsage(toolIDs []string) map[string]int {
	if len(toolIDs) == 0 {
		return nil
	}
	ph := make([]string, len(toolIDs))
	args := make([]interface{}, len(toolIDs))
	for i, id := range toolIDs {
		ph[i] = "?"
		args[i] = id
	}
	q := "SELECT tool_id, select_count FROM tool_usage WHERE tool_id IN (" + strings.Join(ph, ",") + ")"
	rows, err := vs.db.Query(q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	m := make(map[string]int)
	for rows.Next() {
		var id string
		var c int
		if rows.Scan(&id, &c) == nil {
			m[id] = c
		}
	}
	return m
}
