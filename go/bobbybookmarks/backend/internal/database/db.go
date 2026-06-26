package database

import (
	"log"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB
var ArchiveDB *sqlx.DB

func InitDB() {
	var err error
	log.Println("Attempting to connect to database at /data/bookmarks.db...")
	DB, err = sqlx.Connect("sqlite", "/data/bookmarks.db")
	if err != nil {
		log.Printf("Failed to connect to /data/bookmarks.db: %v. Falling back to local bookmarks.db...", err)
		DB, err = sqlx.Connect("sqlite", "bookmarks.db")
		if err != nil {
			log.Fatalf("CRITICAL: Failed to connect to any database: %v", err)
		}
	}
	log.Println("Successfully connected to main database.")

	log.Println("Attempting to connect to archive database...")
	ArchiveDB, err = sqlx.Connect("sqlite", "/data/bookmarks_archive.db")
	if err != nil {
		ArchiveDB, _ = sqlx.Connect("sqlite", "bookmarks_archive.db")
	}
	log.Println("Database initialization complete.")
}

func EnsureSchema() {
	schema := `
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		title TEXT,
		description TEXT,
		page_title TEXT,
		page_description TEXT,
		research_status TEXT DEFAULT 'pending',
		imported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		researched_at DATETIME,
		cluster_id INTEGER,
		source TEXT,
		is_duplicate BOOLEAN DEFAULT 0,
		original_id INTEGER
	);
	CREATE TABLE IF NOT EXISTS clusters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE
	);
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE
	);
	`
	DB.MustExec(schema)
}
