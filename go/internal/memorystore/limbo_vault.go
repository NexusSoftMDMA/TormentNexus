package memorystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

// L4LimboVault stores lost, forgotten, or discarded memories.
// Once a memory enters limbo it can be:
//   - Resurrected (promoted back to L2)
//   - Permanently purged after a configurable TTL
//   - Queried for "what have we forgotten?" analytics
type L4LimboVault struct {
	db *sql.DB
}

const limboSchemaSQL = `
CREATE TABLE IF NOT EXISTS l4_limbo (
    id              TEXT PRIMARY KEY,
    session_id      TEXT NOT NULL DEFAULT '',
    original_tier   TEXT NOT NULL DEFAULT 'l2',
    kind            TEXT NOT NULL DEFAULT '',
    category        TEXT NOT NULL DEFAULT '',
    tags            TEXT NOT NULL DEFAULT '',
    source_url      TEXT NOT NULL DEFAULT '',
    content         TEXT NOT NULL,
    importance      REAL NOT NULL DEFAULT 0.0,
    heat_score      REAL NOT NULL DEFAULT 0.0,
    reason          TEXT NOT NULL DEFAULT 'unknown',
    limbo_entered_at DATETIME NOT NULL DEFAULT (datetime('now')),
    created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    purged_at       DATETIME,
    metadata        TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_limbo_reason   ON l4_limbo(reason);
CREATE INDEX IF NOT EXISTS idx_limbo_entered  ON l4_limbo(limbo_entered_at);
CREATE INDEX IF NOT EXISTS idx_limbo_kind     ON l4_limbo(kind);
`

// LimboReason describes why a memory was sent to limbo.
type LimboReason string

const (
	LimboLost      LimboReason = "lost"      // orphaned / no session reference
	LimboForgotten LimboReason = "forgotten" // manually discarded or timed out
	LimboDiscarded LimboReason = "discarded" // explicitly deleted by user/agent
	LimboDecayed   LimboReason = "decayed"   // heat score dropped to zero
	LimboReplaced  LimboReason = "replaced"  // superseded by a newer memory
)

// NewLimboVault opens or creates the L4 limbo vault.
func NewLimboVault(db *sql.DB) (*L4LimboVault, error) {
	if _, err := db.Exec(limboSchemaSQL); err != nil {
		return nil, fmt.Errorf("limbo schema: %w", err)
	}
	return &L4LimboVault{db: db}, nil
}

// Bury moves a memory record into the limbo vault with a reason.
func (v *L4LimboVault) Bury(ctx context.Context, record controlplane.L2VaultRecord, reason LimboReason) error {
	meta, _ := json.Marshal(map[string]any{
		"original_heat":       record.HeatScore,
		"original_importance": record.Importance,
		"buried_at":           time.Now().UTC().Format(time.RFC3339),
	})
	tier := "l2"
	if record.Type == controlplane.MemoryArchive {
		tier = "l3"
	}
	_, err := v.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO l4_limbo
			(id, session_id, original_tier, kind, category, tags, source_url,
			 content, importance, heat_score, reason, limbo_entered_at, created_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), ?, ?)
	`, record.ID, record.SessionID, tier, record.Kind, record.Category,
		record.Tags, record.SourceURL, record.Content, record.Importance,
		record.HeatScore, string(reason), record.CreatedAt.Format(time.RFC3339), string(meta))
	return err
}

// Resurrect promotes a limbo record back to an L2VaultRecord.
// It does NOT insert it back into L2 — that's the caller's responsibility.
func (v *L4LimboVault) Resurrect(ctx context.Context, id string) (*controlplane.L2VaultRecord, error) {
	row := v.db.QueryRowContext(ctx, `
		SELECT id, session_id, kind, category, tags, source_url,
		       content, importance, heat_score, original_tier, created_at
		FROM l4_limbo WHERE id = ?
	`, id)

	var r controlplane.L2VaultRecord
	var createdAtStr, tier string
	err := row.Scan(&r.ID, &r.SessionID, &r.Kind, &r.Category,
		&r.Tags, &r.SourceURL, &r.Content,
		&r.Importance, &r.HeatScore, &tier, &createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("resurrect find: %w", err)
	}

	r.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	r.LastAccessedAt = time.Now()
	r.Type = controlplane.MemoryLongTerm

	// Boost heat to prevent immediate re-burial
	r.HeatScore = 30.0

	// Delete from limbo
	_, _ = v.db.ExecContext(ctx, `DELETE FROM l4_limbo WHERE id = ?`, id)
	return &r, nil
}

// SearchLimbo returns limbo entries matching a keyword query.
func (v *L4LimboVault) SearchLimbo(ctx context.Context, query string, limit int) ([]controlplane.L2VaultRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	like := "%" + query + "%"
	rows, err := v.db.QueryContext(ctx, `
		SELECT id, session_id, kind, category, tags, source_url,
		       content, importance, heat_score, reason, created_at
		FROM l4_limbo
		WHERE content LIKE ? OR tags LIKE ? OR kind LIKE ? OR reason LIKE ?
		ORDER BY limbo_entered_at DESC
		LIMIT ?
	`, like, like, like, like, limit)
	if err != nil {
		return nil, fmt.Errorf("limbo search: %w", err)
	}
	defer rows.Close()

	var results []controlplane.L2VaultRecord
	for rows.Next() {
		var r controlplane.L2VaultRecord
		var createdAtStr, reason string
		if err := rows.Scan(&r.ID, &r.SessionID, &r.Kind, &r.Category,
			&r.Tags, &r.SourceURL, &r.Content,
			&r.Importance, &r.HeatScore, &reason, &createdAtStr); err != nil {
			continue
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		r.LastAccessedAt = time.Now()
		r.Type = controlplane.MemoryArchive
		results = append(results, r)
	}
	return results, nil
}

// Stats returns counts grouped by reason.
func (v *L4LimboVault) Stats(ctx context.Context) (map[string]int, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT reason, COUNT(*) as cnt
		FROM l4_limbo
		WHERE purged_at IS NULL
		GROUP BY reason
		ORDER BY cnt DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("limbo stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var reason string
		var count int
		if err := rows.Scan(&reason, &count); err != nil {
			continue
		}
		stats[reason] = count
	}
	return stats, nil
}

// PurgeOld permanently removes limbo entries older than the given TTL.
func (v *L4LimboVault) PurgeOld(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan).Format(time.RFC3339)
	res, err := v.db.ExecContext(ctx, `
		UPDATE l4_limbo SET purged_at = datetime('now')
		WHERE purged_at IS NULL AND limbo_entered_at < ?
	`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// HardDelete permanently removes entries that were already purged (soft-delete cleanup).
func (v *L4LimboVault) HardDelete(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan).Format(time.RFC3339)
	res, err := v.db.ExecContext(ctx, `
		DELETE FROM l4_limbo
		WHERE purged_at IS NOT NULL AND purged_at < ?
	`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

var _ = time.Now
