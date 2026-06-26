package memorystore

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const scratchpadSchemaSQL = `
CREATE TABLE IF NOT EXISTS core_memory_scratchpad (
    key          TEXT PRIMARY KEY,
    value        TEXT NOT NULL,
    updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
);
`

// InitScratchpad initializes the scratchpad table schema in the database.
func InitScratchpad(db *sql.DB) error {
	_, err := db.Exec(scratchpadSchemaSQL)
	if err != nil {
		return fmt.Errorf("init scratchpad schema: %w", err)
	}

	// Insert default keys if not present
	_, _ = db.Exec(`INSERT OR IGNORE INTO core_memory_scratchpad (key, value) VALUES ('persona', 'You are a helpful coding assistant.')`)
	_, _ = db.Exec(`INSERT OR IGNORE INTO core_memory_scratchpad (key, value) VALUES ('human', 'No user profile information configured yet.')`)

	return nil
}

// GetScratchpadValue retrieves a core memory value by its key.
func (vs *VectorStore) GetScratchpadValue(ctx context.Context, key string) (string, error) {
	var val string
	err := vs.db.QueryRowContext(ctx, `SELECT value FROM core_memory_scratchpad WHERE key = ?`, key).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

// SetScratchpadValue sets or overwrites a core memory value by key.
func (vs *VectorStore) SetScratchpadValue(ctx context.Context, key, value string) error {
	_, err := vs.db.ExecContext(ctx, `
		INSERT INTO core_memory_scratchpad (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`, key, value, time.Now().UTC().Format("2006-01-02 15:04:05"))
	return err
}

// AppendScratchpadValue appends text to an existing core memory value.
func (vs *VectorStore) AppendScratchpadValue(ctx context.Context, key, contentToAppend string) error {
	current, err := vs.GetScratchpadValue(ctx, key)
	if err != nil {
		return err
	}
	newValue := current
	if newValue != "" {
		newValue += "\n"
	}
	newValue += contentToAppend

	return vs.SetScratchpadValue(ctx, key, newValue)
}
