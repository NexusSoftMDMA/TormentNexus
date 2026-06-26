package models

import (
	"time"
)

type Bookmark struct {
	ID              int       `db:"id" json:"id"`
	URL             string    `db:"url" json:"url"`
	Title           string    `db:"title" json:"title"`
	Description     string    `db:"description" json:"description"` // now empty in main db, in archive
	PageTitle       string    `db:"page_title" json:"page_title"`
	PageDescription string    `db:"page_description" json:"page_description"`
	Status          string    `db:"research_status" json:"research_status"` // pending, running, done, failed
	ImportedAt      time.Time `db:"imported_at" json:"imported_at"`
	ResearchedAt    *time.Time `db:"researched_at" json:"researched_at"`
	ClusterID       *int      `db:"cluster_id" json:"cluster_id"`
	Source          string    `db:"source" json:"source"`
	IsDuplicate     bool      `db:"is_duplicate" json:"is_duplicate"`
	OriginalID      *int      `db:"original_id" json:"original_id"`
}

type Tag struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type Cluster struct {
	ID            int      `db:"id" json:"id"`
	Name          string   `db:"name" json:"name"`
	BookmarkCount int      `json:"bookmark_count"`
	Tags          []string `json:"tags"`
}

type ImportSession struct {
	ID             int       `db:"id" json:"id"`
	SourceType     string    `db:"source_type" json:"source_type"`
	SourceName     string    `db:"source_name" json:"source_name"`
	TotalCount     int       `db:"total_count" json:"total_count"`
	ImportedCount  int       `db:"imported_count" json:"imported_count"`
	DuplicateCount int       `db:"duplicate_count" json:"duplicate_count"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
