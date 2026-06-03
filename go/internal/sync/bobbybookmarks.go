package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Bookmark struct {
	ID              int         `json:"id"`
	URL             string      `json:"url"`
	NormalizedURL   string      `json:"normalized_url"`
	Title           *string     `json:"title"`
	Description     *string     `json:"description"`
	Tags            interface{} `json:"tags"`
	Source          *string     `json:"source"`
	IsDuplicate     bool        `json:"is_duplicate"`
	DuplicateOf     interface{} `json:"duplicate_of"`
	ResearchStatus  *string     `json:"research_status"`
	HTTPStatus      *int        `json:"http_status"`
	PageTitle       *string     `json:"page_title"`
	PageDescription *string     `json:"page_description"`
	FaviconURL      *string     `json:"favicon_url"`
	ResearchedAt    *string     `json:"researched_at"`
	ClusterID       interface{} `json:"cluster_id"`
	ImportSessionID *int        `json:"import_session_id"`
}

type SyncReport struct {
	Source   string   `json:"source"`
	Fetched  int      `json:"fetched"`
	Upserted int      `json:"upserted"`
	Pages    int      `json:"pages"`
	Errors   []string `json:"errors"`
	BaseURL  string   `json:"baseUrl"`
}

func coerceTags(tags interface{}) string {
	if tags == nil {
		return "[]"
	}
	// Simplified coercion for now: just marshal back to JSON or parse slice
	b, _ := json.Marshal(tags)
	return string(b)
}


func isGarbageExtraction(title *string, description *string) bool {
	check := func(s *string) bool {
		if s == nil {
			return false
		}
		lower := strings.ToLower(*s)
		return strings.Contains(lower, "automated discovery") ||
			strings.Contains(lower, "heuristic mapping")
	}
	return check(title) || check(description)
}

func SyncBobbyBookmarks(ctx context.Context, dbPath string, baseURL string, perPage int, includeDuplicates bool, includeResearched bool) (*SyncReport, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	report := &SyncReport{
		Source:  "bobbybookmarks",
		BaseURL: baseURL,
		Errors:  []string{},
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	client := &http.Client{Timeout: 15 * time.Second}

	for page := 1; page <= 100; page++ {
		reqURL, _ := url.Parse(fmt.Sprintf("%s/api/bookmarks", baseURL))
		q := reqURL.Query()
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", strconv.Itoa(perPage))
		if includeDuplicates {
			q.Set("show_duplicates", "true")
		}
		if !includeResearched {
			q.Set("research_status", "pending")
		}
		reqURL.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("page %d req error: %v", page, err))
			break
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "TormentNexus/BobbyBookmarks-Go-Adapter")

		resp, err := client.Do(req)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("page %d fetch error: %v", page, err))
			break
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			report.Errors = append(report.Errors, fmt.Sprintf("page %d HTTP %d", page, resp.StatusCode))
			break
		}

		var payload struct {
			Data []Bookmark `json:"data"`
			// Handle fallback structure
			Items []Bookmark `json:"items"`
		}
		
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if decodeErr != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("page %d decode error: %v", page, decodeErr))
			break
		}

		bookmarks := payload.Data
		if len(bookmarks) == 0 {
			bookmarks = payload.Items
		}

		if len(bookmarks) == 0 {
			break // Reached end
		}

		report.Pages++
		report.Fetched += len(bookmarks)

		// Begin transaction for bulk upsert
		tx, err := db.Begin()
		if err != nil {
			report.Errors = append(report.Errors, "tx begin error: "+err.Error())
			break
		}

		stmt, err := tx.Prepare(`
			INSERT INTO links_backlog (
				uuid, url, normalized_url, title, description, tags, source, 
				is_duplicate, duplicate_of, research_status, http_status, 
				page_title, page_description, favicon_url, cluster_id, 
				bobbybookmarks_bookmark_id, import_session_id, synced_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(url) DO UPDATE SET
				normalized_url = excluded.normalized_url,
				title = coalesce(excluded.title, links_backlog.title),
				description = coalesce(excluded.description, links_backlog.description),
				tags = excluded.tags,
				is_duplicate = excluded.is_duplicate,
				duplicate_of = excluded.duplicate_of,
				research_status = excluded.research_status,
				http_status = coalesce(excluded.http_status, links_backlog.http_status),
				page_title = coalesce(excluded.page_title, links_backlog.page_title),
				page_description = coalesce(excluded.page_description, links_backlog.page_description),
				favicon_url = coalesce(excluded.favicon_url, links_backlog.favicon_url),
				cluster_id = excluded.cluster_id,
				bobbybookmarks_bookmark_id = excluded.bobbybookmarks_bookmark_id,
				import_session_id = excluded.import_session_id,
				synced_at = excluded.synced_at,
				updated_at = excluded.updated_at
		`)

		if err != nil {
			tx.Rollback()
			report.Errors = append(report.Errors, "stmt prepare error: "+err.Error())
			break
		}

		now := time.Now().UnixMilli()

		for _, bm := range bookmarks {
			if strings.TrimSpace(bm.URL) == "" {
				continue
			}

			normURL := bm.NormalizedURL
			if normURL == "" {
				normURL = bm.URL
			}

			dupOf := ""
			if bm.DuplicateOf != nil {
				dupOf = fmt.Sprintf("%v", bm.DuplicateOf)
			}
			
			clusterID := ""
			if bm.ClusterID != nil {
				clusterID = fmt.Sprintf("%v", bm.ClusterID)
			}

			rStatus := "pending"
			if bm.ResearchStatus != nil {
				rStatus = *bm.ResearchStatus
			}

			if isGarbageExtraction(bm.Title, bm.Description) {
				report.Errors = append(report.Errors, fmt.Sprintf("Rejected garbage entry: %s", bm.URL))
				continue
			}

			uid := uuid.New().String()


			_, err = stmt.Exec(
				uid, bm.URL, normURL, bm.Title, bm.Description, coerceTags(bm.Tags), "bobbybookmarks",
				bm.IsDuplicate, dupOf, rStatus, bm.HTTPStatus,
				bm.PageTitle, bm.PageDescription, bm.FaviconURL, clusterID,
				bm.ID, bm.ImportSessionID, now, now, now,
			)
			if err == nil {
				report.Upserted++
			}
		}

		if err := stmt.Close(); err != nil {
			tx.Rollback()
			report.Errors = append(report.Errors, "stmt close error: "+err.Error())
			break
		}
		if err := tx.Commit(); err != nil {
			report.Errors = append(report.Errors, "tx commit error: "+err.Error())
			break
		}

		if len(bookmarks) < perPage {
			break
		}
	}

	return report, nil
}
