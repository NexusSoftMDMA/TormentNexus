package memorystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
	"github.com/tormentnexushq/tormentnexus-go/internal/gossip"
)

// InitGossipStore initializes the schemas required for gossip metadata tracking.
func InitGossipStore(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS gossip_clock (
		node_id TEXT PRIMARY KEY,
		version INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS gossip_metadata (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		version INTEGER NOT NULL,
		origin TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		hash TEXT NOT NULL
	);
	`)
	return err
}

// GossipStoreAdapter implements the gossip.StateStore interface.
type GossipStoreAdapter struct {
	vs *VectorStore
}

func NewGossipStoreAdapter(vs *VectorStore) *GossipStoreAdapter {
	_ = InitGossipStore(vs.db)
	return &GossipStoreAdapter{vs: vs}
}

// GetDigest returns version digests for all local entries.
func (a *GossipStoreAdapter) GetDigest(ctx context.Context) ([]gossip.DigestEntry, error) {
	rows, err := a.vs.db.QueryContext(ctx, `SELECT id, version, hash FROM gossip_metadata`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var digests []gossip.DigestEntry
	for rows.Next() {
		var de gossip.DigestEntry
		if err := rows.Scan(&de.ID, &de.Version, &de.Hash); err != nil {
			return nil, err
		}
		digests = append(digests, de)
	}
	return digests, nil
}

// GetEntries returns full entries for the given IDs.
func (a *GossipStoreAdapter) GetEntries(ctx context.Context, ids []string) ([]gossip.StateEntry, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var entries []gossip.StateEntry
	for _, id := range ids {
		var se gossip.StateEntry
		// Fetch metadata first
		err := a.vs.db.QueryRowContext(ctx, `
			SELECT id, type, version, origin, timestamp, hash 
			FROM gossip_metadata WHERE id = ?`, id).Scan(&se.ID, &se.Type, &se.Version, &se.Origin, &se.Timestamp, &se.Hash)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}

		if se.Type == "memory" {
			// Fetch actual content from l2_vault
			var content string
			err := a.vs.db.QueryRowContext(ctx, `SELECT content FROM l2_vault WHERE id = ?`, id).Scan(&content)
			if err != nil {
				if err == sql.ErrNoRows {
					continue
				}
				return nil, err
			}
			se.Content = content
		}
		entries = append(entries, se)
	}
	return entries, nil
}

// Merge applies remote entries, using last-write-wins with vector clock.
func (a *GossipStoreAdapter) Merge(ctx context.Context, entries []gossip.StateEntry) (int, error) {
	accepted := 0
	for _, entry := range entries {
		var localVersion uint64
		err := a.vs.db.QueryRowContext(ctx, `SELECT version FROM gossip_metadata WHERE id = ?`, entry.ID).Scan(&localVersion)
		if err != nil && err != sql.ErrNoRows {
			return accepted, err
		}

		if err == sql.ErrNoRows || entry.Version > localVersion {
			// Save actual memory first
			if entry.Type == "memory" {
				var record controlplane.L2VaultRecord
				if err := json.Unmarshal([]byte(entry.Content), &record); err != nil {
					// Fallback: use simple string as content
					record = controlplane.L2VaultRecord{
						ID:        entry.ID,
						SessionID: "gossip-sync",
						Type:      controlplane.MemoryWorking,
						Kind:      "fact",
						Category:  "gossip",
						Content:   entry.Content,
						CreatedAt: time.UnixMilli(entry.Timestamp),
					}
				}
				if err := a.vs.Commit(ctx, record); err != nil {
					return accepted, err
				}
			}

			// Update metadata
			_, err = a.vs.db.ExecContext(ctx, `
				INSERT INTO gossip_metadata (id, type, version, origin, timestamp, hash)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					version = excluded.version,
					origin = excluded.origin,
					timestamp = excluded.timestamp,
					hash = excluded.hash
			`, entry.ID, entry.Type, entry.Version, entry.Origin, entry.Timestamp, entry.Hash)
			if err != nil {
				return accepted, err
			}
			accepted++
		}
	}
	return accepted, nil
}

// LocalClock returns the current vector clock.
func (a *GossipStoreAdapter) LocalClock(ctx context.Context) (gossip.VectorClock, error) {
	rows, err := a.vs.db.QueryContext(ctx, `SELECT node_id, version FROM gossip_clock`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clock := gossip.VectorClock{}
	for rows.Next() {
		var nodeID string
		var version uint64
		if err := rows.Scan(&nodeID, &version); err != nil {
			return nil, err
		}
		clock[nodeID] = version
	}
	return clock, nil
}

// IncrementClock increments the clock for this node.
func (a *GossipStoreAdapter) IncrementClock(ctx context.Context) (uint64, error) {
	var current uint64
	err := a.vs.db.QueryRowContext(ctx, `SELECT version FROM gossip_clock WHERE node_id = 'local'`).Scan(&current)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	next := current + 1
	_, err = a.vs.db.ExecContext(ctx, `
		INSERT INTO gossip_clock (node_id, version) VALUES ('local', ?)
		ON CONFLICT(node_id) DO UPDATE SET version = excluded.version
	`, next)
	if err != nil {
		return 0, err
	}
	return next, nil
}
