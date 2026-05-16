package memorystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/borghq/borg-go/internal/controlplane"
	_ "modernc.org/sqlite"
)

type VectorStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewVectorStore(dbPath string) (*VectorStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if dbPath != ":memory:" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set WAL mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set busy timeout: %w", err)
		}
	}

	if _, err := db.Exec(controlplane.VectorSchemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init vector schema: %w", err)
	}

	return &VectorStore{db: db}, nil
}

func (s *VectorStore) Close() error {
	return s.db.Close()
}

func (s *VectorStore) Commit(ctx context.Context, entry controlplane.L2VaultRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.HeatScore == 0 {
		entry.HeatScore = 50.0
	}
	if entry.LastAccessedAt.IsZero() {
		entry.LastAccessedAt = time.Now()
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO l2_vault (id, session_id, memory_type, content, importance, heat_score, last_accessed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content = excluded.content,
			importance = excluded.importance,
			heat_score = excluded.heat_score,
			last_accessed_at = excluded.last_accessed_at,
			created_at = excluded.created_at
	`, entry.ID, entry.SessionID, string(entry.Type), entry.Content, entry.Importance, entry.HeatScore, entry.LastAccessedAt, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("memorystore commit insert: %w", err)
	}

	if len(entry.Embedding) > 0 {
		embeddingJSON, _ := json.Marshal(entry.Embedding)
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO vec_l2_vault (rowid, embedding)
			SELECT rowid, ? FROM l2_vault WHERE id = ?
		`, string(embeddingJSON), entry.ID)
		if err != nil {
			return fmt.Errorf("memorystore commit embedding: %w", err)
		}
	}
	return nil
}

func (s *VectorStore) SemanticSearch(ctx context.Context, query string, limit int) ([]controlplane.L2VaultRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queryStr := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, memory_type, content, importance, heat_score, last_accessed_at, created_at
		FROM l2_vault
		WHERE content LIKE ? AND memory_type != 'archive'
		ORDER BY importance DESC, heat_score DESC, created_at DESC
		LIMIT ?
	`, queryStr, limit)
	if err != nil {
		return nil, fmt.Errorf("memorystore search: %w", err)
	}
	defer rows.Close()

	var results []controlplane.L2VaultRecord
	for rows.Next() {
		var r controlplane.L2VaultRecord
		var mType string
		if err := rows.Scan(&r.ID, &r.SessionID, &mType, &r.Content, &r.Importance, &r.HeatScore, &r.LastAccessedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Type = controlplane.MemoryType(mType)
		results = append(results, r)
	}

	// Update heat and last_accessed_at for hits
	for _, r := range results {
		s.incrementHeatLocked(ctx, r.ID)
	}

	return results, nil
}

func (s *VectorStore) incrementHeatLocked(ctx context.Context, id string) {
	_, _ = s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET heat_score = MIN(100.0, heat_score + 10.0),
		    last_accessed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id)
}

func (s *VectorStore) ApplyDecay(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// heat_score = heat_score * exp(-0.0288 * hours_since_access)
	_, err := s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET heat_score = heat_score * exp(-0.0288 * (julianday('now') - julianday(last_accessed_at)) * 24.0)
		WHERE memory_type != 'archive'
	`)
	if err != nil {
		return fmt.Errorf("apply decay: %w", err)
	}

	// Promote: Working memories with a heat > 80 move to long_term
	_, err = s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET memory_type = 'long_term'
		WHERE heat_score > 80.0 AND memory_type = 'working'
	`)
	if err != nil {
		return fmt.Errorf("promotion: %w", err)
	}

	// Demote: long_term memories with a heat < 20 move to the archive (L3)
	_, err = s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET memory_type = 'archive'
		WHERE heat_score < 20.0 AND memory_type = 'long_term'
	`)

	return err
}
