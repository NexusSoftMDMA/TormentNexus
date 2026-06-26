package memorystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

// RelationStore implements GraphRAG semantic/structural edge persistence.
// It stores typed, weighted relations between memory records and supports
// graph traversal queries for context retrieval.
type RelationStore struct {
	db *sql.DB
}

const relationSchemaSQL = `
CREATE TABLE IF NOT EXISTS memory_relations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id   TEXT NOT NULL,
    target_id   TEXT NOT NULL,
    rel_type    TEXT NOT NULL,
    weight      REAL NOT NULL DEFAULT 1.0,
    metadata    TEXT NOT NULL DEFAULT '{}',
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(source_id, target_id, rel_type)
);

CREATE INDEX IF NOT EXISTS idx_relations_source ON memory_relations(source_id);
CREATE INDEX IF NOT EXISTS idx_relations_target ON memory_relations(target_id);
CREATE INDEX IF NOT EXISTS idx_relations_type   ON memory_relations(rel_type);

-- Full-text search on relation metadata
CREATE VIRTUAL TABLE IF NOT EXISTS memory_relations_fts USING fts5(
    relation_id UNINDEXED,
    metadata,
    content='memory_relations',
    content_rowid='id',
    tokenize='porter unicode61'
);

CREATE TRIGGER IF NOT EXISTS relations_ai AFTER INSERT ON memory_relations BEGIN
    INSERT INTO memory_relations_fts(rowid, relation_id, metadata)
    VALUES (new.id, new.id, new.metadata);
END;

CREATE TRIGGER IF NOT EXISTS relations_ad AFTER DELETE ON memory_relations BEGIN
    INSERT INTO memory_relations_fts(memory_relations_fts, rowid, relation_id, metadata)
    VALUES ('delete', old.id, old.id, old.metadata);
END;
`

// NewRelationStore opens (or creates) the relation store.
func NewRelationStore(db *sql.DB) (*RelationStore, error) {
	if _, err := db.Exec(relationSchemaSQL); err != nil {
		return nil, fmt.Errorf("relation schema: %w", err)
	}
	return &RelationStore{db: db}, nil
}

// AddRelation creates a typed edge between two memory records.
// If the edge already exists, it updates the weight and timestamp.
func (rs *RelationStore) AddRelation(ctx context.Context, sourceID, targetID, relType string, weight float64, metadata map[string]any) error {
	meta, _ := json.Marshal(metadata)
	if meta == nil {
		meta = []byte("{}")
	}
	_, err := rs.db.ExecContext(ctx, `
		INSERT INTO memory_relations (source_id, target_id, rel_type, weight, metadata, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(source_id, target_id, rel_type) DO UPDATE SET
			weight = excluded.weight,
			metadata = excluded.metadata,
			updated_at = datetime('now')
	`, sourceID, targetID, relType, weight, string(meta))
	return err
}

// RemoveRelation deletes a specific edge.
func (rs *RelationStore) RemoveRelation(ctx context.Context, sourceID, targetID, relType string) error {
	_, err := rs.db.ExecContext(ctx, `
		DELETE FROM memory_relations
		WHERE source_id = ? AND target_id = ? AND rel_type = ?
	`, sourceID, targetID, relType)
	return err
}

// GetRelations returns all outgoing relations from a memory record.
func (rs *RelationStore) GetRelations(ctx context.Context, sourceID string) ([]controlplane.L2Relation, error) {
	rows, err := rs.db.QueryContext(ctx, `
		SELECT source_id, target_id, rel_type, weight
		FROM memory_relations
		WHERE source_id = ?
		ORDER BY weight DESC
	`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("get relations: %w", err)
	}
	defer rows.Close()

	var rels []controlplane.L2Relation
	for rows.Next() {
		var r controlplane.L2Relation
		if err := rows.Scan(&r.SourceID, &r.TargetID, &r.RelationType, &r.Weight); err != nil {
			continue
		}
		rels = append(rels, r)
	}
	return rels, nil
}

// GetInbound returns all incoming relations to a memory record (reverse lookup).
func (rs *RelationStore) GetInbound(ctx context.Context, targetID string) ([]controlplane.L2Relation, error) {
	rows, err := rs.db.QueryContext(ctx, `
		SELECT source_id, target_id, rel_type, weight
		FROM memory_relations
		WHERE target_id = ?
		ORDER BY weight DESC
	`, targetID)
	if err != nil {
		return nil, fmt.Errorf("get inbound: %w", err)
	}
	defer rows.Close()

	var rels []controlplane.L2Relation
	for rows.Next() {
		var r controlplane.L2Relation
		if err := rows.Scan(&r.SourceID, &r.TargetID, &r.RelationType, &r.Weight); err != nil {
			continue
		}
		rels = append(rels, r)
	}
	return rels, nil
}

// Traverse performs a breadth-first graph traversal from a starting memory,
// returning all reachable records up to the specified depth.
func (rs *RelationStore) Traverse(ctx context.Context, startID string, maxDepth int, minWeight float64) ([]TraversalNode, error) {
	if maxDepth <= 0 || maxDepth > 5 {
		maxDepth = 3
	}

	type node struct {
		id     string
		depth  int
		rel    string
		weight float64
	}

	visited := map[string]bool{startID: true}
	queue := []node{{id: startID, depth: 0}}
	var results []TraversalNode

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth > 0 {
			results = append(results, TraversalNode{
				ID:      current.id,
				Depth:   current.depth,
				RelType: current.rel,
				Weight:  current.weight,
			})
		}

		if current.depth >= maxDepth {
			continue
		}

		rows, err := rs.db.QueryContext(ctx, `
			SELECT target_id, rel_type, weight
			FROM memory_relations
			WHERE source_id = ? AND weight >= ?
			ORDER BY weight DESC
			LIMIT 20
		`, current.id, minWeight)
		if err != nil {
			return nil, fmt.Errorf("traverse query: %w", err)
		}
		for rows.Next() {
			var targetID, relType string
			var weight float64
			if err := rows.Scan(&targetID, &relType, &weight); err != nil {
				continue
			}
			if !visited[targetID] {
				visited[targetID] = true
				queue = append(queue, node{id: targetID, depth: current.depth + 1, rel: relType, weight: weight})
			}
		}
		rows.Close()
	}

	return results, nil
}

// TraversalNode represents a node discovered during graph traversal.
type TraversalNode struct {
	ID      string  `json:"id"`
	Depth   int     `json:"depth"`
	RelType string  `json:"relation_type"`
	Weight  float64 `json:"weight"`
}

// GetStats returns relation counts by type.
func (rs *RelationStore) GetStats(ctx context.Context) (map[string]int, error) {
	rows, err := rs.db.QueryContext(ctx, `
		SELECT rel_type, COUNT(*) as cnt
		FROM memory_relations
		GROUP BY rel_type
		ORDER BY cnt DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("relation stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var relType string
		var count int
		if err := rows.Scan(&relType, &count); err != nil {
			continue
		}
		stats[relType] = count
	}
	return stats, nil
}

var _ = time.Now
