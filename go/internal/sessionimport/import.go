package sessionimport

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type ImportedSession struct {
	ID                string `json:"id"`
	SourceTool        string `json:"sourceTool"`
	SourcePath        string `json:"sourcePath"`
	ExternalSessionID string `json:"externalSessionId"`
	Title             string `json:"title"`
	SessionFormat     string `json:"sessionFormat"`
	Transcript        string `json:"transcript"`
}

func ImportSession(dbPath string, session ImportedSession) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open tormentnexus.db: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UnixMilli()

	// Check if already exists based on SourcePath
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM imported_sessions WHERE source_path = ?)", session.SourcePath).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		_, err = tx.Exec(`
			UPDATE imported_sessions 
			SET transcript = ?, updated_at = ?
			WHERE source_path = ?
		`, session.Transcript, now, session.SourcePath)
	} else {
		newUUID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO imported_sessions (
				uuid, source_tool, source_path, external_session_id, title, 
				session_format, transcript, discovered_at, imported_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, newUUID, session.SourceTool, session.SourcePath, session.ExternalSessionID, session.Title,
			session.SessionFormat, session.Transcript, now, now, now, now)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
