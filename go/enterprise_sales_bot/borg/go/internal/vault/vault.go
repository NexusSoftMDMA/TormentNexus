package vault

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type EmbedFunc func(text string) ([]float32, error)

type Vault struct {
	db            *sql.DB
	mu            sync.RWMutex
	stmtStore     *sql.Stmt
	stmtStoreEmbed *sql.Stmt
	stmtGetEntry  *sql.Stmt
}

func OpenVault(path string) (*Vault, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil { return nil, fmt.Errorf("open vault %s: %w", path, err) }
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA cache_size=-64000")
	if err := InitL2Schema(db); err != nil { db.Close(); return nil, fmt.Errorf("init L2: %w", err) }
	v := &Vault{db: db}
	if err := v.prepareStatements(); err != nil { db.Close(); return nil, err }
	return v, nil
}

func (v *Vault) prepareStatements() error {
	var err error
	v.stmtStore, err = v.db.Prepare(`INSERT INTO l2_entries (id, memory_type, session_id, agent_name, project, content, tags, importance, source_models) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil { return fmt.Errorf("prepare store: %w", err) }
	v.stmtStoreEmbed, err = v.db.Prepare(`INSERT INTO l2_embeddings (entry_id, model_name, dimension, vector) VALUES (?, ?, ?, ?) ON CONFLICT(entry_id, model_name) DO UPDATE SET vector=excluded.vector, dimension=excluded.dimension`)
	if err != nil { return fmt.Errorf("prepare store embed: %w", err) }
	v.stmtGetEntry, err = v.db.Prepare(`SELECT id, memory_type, session_id, agent_name, project, content, tags, importance, source_models, created_at FROM l2_entries WHERE id = ?`)
	if err != nil { return fmt.Errorf("prepare get: %w", err) }
	return nil
}

func (v *Vault) Close() error {
	if v.stmtStore != nil { v.stmtStore.Close() }
	if v.stmtStoreEmbed != nil { v.stmtStoreEmbed.Close() }
	if v.stmtGetEntry != nil { v.stmtGetEntry.Close() }
	return v.db.Close()
}

func (v *Vault) Commit(entry VaultEntry) (string, error) {
	v.mu.Lock(); defer v.mu.Unlock()
	id := entry.ID
	if id == "" { id = uuid.New().String() }
	_, err := v.stmtStore.Exec(id, entry.MemoryType, entry.SessionID, entry.AgentName, entry.Project, entry.Content, entry.Tags, entry.Importance, entry.SourceModels)
	if err != nil { return "", fmt.Errorf("commit: %w", err) }
	return id, nil
}

func (v *Vault) CommitWithEmbedding(entry VaultEntry, modelName string, vec []float32) (string, error) {
	id, err := v.Commit(entry)
	if err != nil { return "", err }
	v.mu.Lock(); defer v.mu.Unlock()
	_, err = v.stmtStoreEmbed.Exec(id, modelName, len(vec), encodeVaultVector(vec))
	if err != nil { return "", fmt.Errorf("commit embed: %w", err) }
	return id, nil
}

func (v *Vault) CommitSession(session *Session, entries []ScratchEntry, heuristicSummary string, embedFn EmbedFunc) error {
	var sb strings.Builder
	for _, e := range entries {
		if e.ToolName != "" {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s | tool=%s(%s) -> %s\n", e.CreatedAt.Format(time.RFC3339), e.Role, e.Content, e.ToolName, e.ToolArgs, e.ToolResult))
		} else {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", e.CreatedAt.Format(time.RFC3339), e.Role, e.Content))
		}
	}
	rawContent := sb.String()
	rawEntry := VaultEntry{MemoryType: MemoryRaw, SessionID: session.ID, AgentName: session.AgentName, Project: session.Project, Content: rawContent, Importance: 0.3}
	rawID, err := v.Commit(rawEntry)
	if err != nil { return fmt.Errorf("commit raw: %w", err) }
	if embedFn != nil {
		go func() {
			vec, err := embedFn(rawContent)
			if err == nil && len(vec) > 0 {
				v.mu.Lock()
				v.stmtStoreEmbed.Exec(rawID, "all-MiniLM-L6-v2", len(vec), encodeVaultVector(vec))
				v.mu.Unlock()
			}
		}()
	}
	if heuristicSummary != "" {
		hEntry := VaultEntry{MemoryType: MemoryHeuristic, SessionID: session.ID, AgentName: session.AgentName, Project: session.Project, Content: heuristicSummary, Importance: 0.7, Tags: "lessons-learned,auto-summarized"}
		hID, err := v.Commit(hEntry)
		if err != nil { return fmt.Errorf("commit heuristic: %w", err) }
		if embedFn != nil {
			go func() {
				vec, err := embedFn(heuristicSummary)
				if err == nil && len(vec) > 0 {
					v.mu.Lock()
					v.stmtStoreEmbed.Exec(hID, "all-MiniLM-L6-v2", len(vec), encodeVaultVector(vec))
					v.mu.Unlock()
				}
			}()
		}
	}
	return nil
}

func (v *Vault) Get(id string) (*VaultEntry, error) {
	v.mu.RLock(); defer v.mu.RUnlock()
	var e VaultEntry; var createdAt string
	err := v.stmtGetEntry.QueryRow(id).Scan(&e.ID, &e.MemoryType, &e.SessionID, &e.AgentName, &e.Project, &e.Content, &e.Tags, &e.Importance, &e.SourceModels, &createdAt)
	if err == sql.ErrNoRows { return nil, nil }
	if err != nil { return nil, err }
	e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &e, nil
}

func (v *Vault) Search(queryVec []float32, topK int, minScore float64, favorHeuristics bool) ([]ContextHit, error) {
	v.mu.RLock(); defer v.mu.RUnlock()
	if topK <= 0 { topK = 5 }
	if minScore <= 0 { minScore = 0.3 }
	rows, err := v.db.Query(`SELECT emb.entry_id, emb.vector, emb.dimension, ent.id, ent.memory_type, ent.session_id, ent.agent_name, ent.project, ent.content, ent.tags, ent.importance, ent.source_models, ent.created_at FROM l2_embeddings emb JOIN l2_entries ent ON ent.id = emb.entry_id WHERE emb.model_name = 'all-MiniLM-L6-v2'`)
	if err != nil { return nil, fmt.Errorf("search: %w", err) }
	defer rows.Close()
	type scored struct { entry VaultEntry; score float64 }
	var results []scored
	for rows.Next() {
		var entryID string; var blob []byte; var dim int; var e VaultEntry; var createdAt string
		if rows.Scan(&entryID, &blob, &dim, &e.ID, &e.MemoryType, &e.SessionID, &e.AgentName, &e.Project, &e.Content, &e.Tags, &e.Importance, &e.SourceModels, &createdAt) != nil { continue }
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		vec := decodeVaultVector(blob, dim)
		score := cosineSim(queryVec, vec)
		if favorHeuristics && e.MemoryType == MemoryHeuristic { score *= 1.2 }
		score *= (0.5 + 0.5*e.Importance)
		if score >= minScore { results = append(results, scored{entry: e, score: score}) }
	}
	sort.Slice(results, func(i, j int) bool { return results[i].score > results[j].score })
	if len(results) > topK { results = results[:topK] }
	hits := make([]ContextHit, len(results))
	for i := range hits {
		hits[i] = ContextHit{Entry: results[i].entry, Score: results[i].score, Rank: i+1}
	}
	return hits, nil
}

func (v *Vault) RecentEntries(n int, memType MemoryType) ([]VaultEntry, error) {
	v.mu.RLock(); defer v.mu.RUnlock()
	q := `SELECT id, memory_type, session_id, agent_name, project, content, tags, importance, source_models, created_at FROM l2_entries`
	args := []interface{}{}
	if memType != "" { q += " WHERE memory_type = ?"; args = append(args, memType) }
	q += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, n)
	rows, err := v.db.Query(q, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var entries []VaultEntry
	for rows.Next() {
		var e VaultEntry; var createdAt string
		if rows.Scan(&e.ID, &e.MemoryType, &e.SessionID, &e.AgentName, &e.Project, &e.Content, &e.Tags, &e.Importance, &e.SourceModels, &createdAt) != nil { continue }
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, e)
	}
	return entries, nil
}

func encodeVaultVector(v []float32) []byte {
	buf := make([]byte, len(v)*4)
	for i, f := range v { binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f)) }
	return buf
}

func decodeVaultVector(buf []byte, dim int) []float32 {
	if len(buf) < dim*4 { dim = len(buf)/4 }
	v := make([]float32, dim)
	for i := 0; i < dim; i++ { v[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf[i*4:])) }
	return v
}

func cosineSim(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 { return 0 }
	var dot, nA, nB float64
	for i := range a { af := float64(a[i]); bf := float64(b[i]); dot += af*bf; nA += af*af; nB += bf*bf }
	if nA == 0 || nB == 0 { return 0 }
	return dot / (math.Sqrt(nA) * math.Sqrt(nB))
}
