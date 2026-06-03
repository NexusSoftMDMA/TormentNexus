package hsync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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
	switch v := tags.(type) {
	case string:
		if v == "" {
			return "[]"
		}
		parts := strings.Split(v, ",")
		var result []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		b, _ := json.Marshal(result)
		return string(b)
	case []interface{}:
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
		}
		b, _ := json.Marshal(result)
		return string(b)
	case []string:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func NormalizeBookmarkURL(rawURL string) string {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return ""
	}

	candidate := trimmed
	if !strings.Contains(trimmed, "://") {
		candidate = "https://" + trimmed
	}

	u, err := url.Parse(candidate)
	if err != nil {
		return strings.ToLower(trimmed)
	}

	blockedParams := map[string]bool{
		"utm_source":   true,
		"utm_medium":   true,
		"utm_campaign": true,
		"utm_term":     true,
		"utm_content":  true,
		"utm_id":       true,
		"utm_reader":   true,
		"utm_name":     true,
		"utm_cid":      true,
		"fbclid":       true,
		"gclid":        true,
		"gclsrc":       true,
		"dclid":        true,
		"msclkid":      true,
		"adrefer":      true,
		"ref":          true,
		"source":       true,
		"mc_cid":       true,
		"mc_eid":       true,
		"zanpid":       true,
		"openid":       true,
		"_ga":          true,
		"_gid":         true,
		"igshid":       true,
		"yclid":        true,
		"twclid":       true,
		"li_fat_id":    true,
		"epik":         true,
		"rdid":         true,
		"ttclid":       true,
		"wbraid":       true,
		"gbraid":       true,
		"srsltid":      true,
	}

	defaultPorts := map[string]string{
		"http:":  "80",
		"https:": "443",
		"ftp:":   "21",
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if port := u.Port(); port != "" {
		if defaultPort, ok := defaultPorts[u.Scheme+":"]; ok && port == defaultPort {
			u.Host = strings.Split(u.Host, ":")[0]
		}
	}

	u.Path = strings.ToLower(u.Path)
	if u.Path != "/" && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimRight(u.Path, "/")
	}
	if u.Path == "" {
		u.Path = "/"
	}

	q := u.Query()
	var keys []string
	for k := range q {
		lowerK := strings.ToLower(k)
		if blockedParams[lowerK] || strings.HasPrefix(lowerK, "utm_") {
			q.Del(k)
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	newQ := url.Values{}
	for _, k := range keys {
		vals := q[k]
		for _, v := range vals {
			newQ.Add(strings.ToLower(k), v)
		}
	}
	u.RawQuery = newQ.Encode()
	u.Fragment = ""

	return u.String()
}

func upsertBookmarks(tx *sql.Tx, bookmarks []Bookmark) (int, error) {
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
		return 0, err
	}
	defer stmt.Close()

	now := time.Now().UnixMilli()
	upserted := 0

	for _, bm := range bookmarks {
		if strings.TrimSpace(bm.URL) == "" {
			continue
		}

		normURL := bm.NormalizedURL
		if normURL == "" {
			normURL = NormalizeBookmarkURL(bm.URL)
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

		uid := uuid.New().String()

		_, err = stmt.Exec(
			uid, bm.URL, normURL, bm.Title, bm.Description, coerceTags(bm.Tags), "bobbybookmarks",
			bm.IsDuplicate, dupOf, rStatus, bm.HTTPStatus,
			bm.PageTitle, bm.PageDescription, bm.FaviconURL, clusterID,
			bm.ID, bm.ImportSessionID, now, now, now,
		)
		if err == nil {
			upserted++
		}
	}

	return upserted, nil
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
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
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

		upserted, err := upsertBookmarks(tx, bookmarks)
		if err != nil {
			tx.Rollback()
			report.Errors = append(report.Errors, "upsert error: "+err.Error())
			break
		}
		report.Upserted += upserted

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

func SyncBobbyBookmarksFromText(ctx context.Context, destDbPath string, textFilePath string) (*SyncReport, error) {
	report := &SyncReport{
		Source:  "bobbybookmarks-text",
		BaseURL: "file://" + textFilePath,
		Errors:  []string{},
	}

	if _, err := os.Stat(textFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("text file not found: %s", textFilePath)
	}

	data, err := os.ReadFile(textFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var bookmarks []Bookmark
	seen := make(map[string]bool)

	for i, line := range lines {
		rawURL := strings.TrimSpace(line)
		if rawURL == "" || strings.HasPrefix(rawURL, "#") {
			continue
		}

		normURL := NormalizeBookmarkURL(rawURL)
		if normURL == "" {
			continue
		}

		if seen[normURL] {
			continue
		}
		seen[normURL] = true

		title := fmt.Sprintf("Bookmark from %s line %d", filepath.Base(textFilePath), i+1)
		bookmarks = append(bookmarks, Bookmark{
			ID:            i + 1,
			URL:           rawURL,
			NormalizedURL: normURL,
			Title:         &title,
		})
	}

	report.Fetched = len(bookmarks)

	db, err := sql.Open("sqlite", destDbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open destination database: %w", err)
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("tx begin error: %w", err)
	}

	upserted, err := upsertBookmarks(tx, bookmarks)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("upsert error: %w", err)
	}
	report.Upserted = upserted

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx commit error: %w", err)
	}

	return report, nil
}

func SyncBobbyBookmarksLocal(ctx context.Context, destDbPath string, sourceDbPath string) (*SyncReport, error) {
	report := &SyncReport{
		Source:  "bobbybookmarks-local",
		BaseURL: "local://" + sourceDbPath,
		Errors:  []string{},
	}

	if _, err := os.Stat(sourceDbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source database not found: %s", sourceDbPath)
	}

	sourceDb, err := sql.Open("sqlite", sourceDbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source database: %w", err)
	}
	sourceDb.Exec("PRAGMA journal_mode=WAL")
	sourceDb.Exec("PRAGMA busy_timeout=5000")
	defer sourceDb.Close()

	destDb, err := sql.Open("sqlite", destDbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open destination database: %w", err)
	}
	destDb.Exec("PRAGMA journal_mode=WAL")
	destDb.Exec("PRAGMA busy_timeout=5000")
	defer destDb.Close()

	rows, err := sourceDb.QueryContext(ctx, "SELECT id, url, short_description, long_description, tags, research_level FROM bookmarks")
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		var shortDesc, longDesc, tags sql.NullString
		var researchLevel sql.NullString
		if err := rows.Scan(&b.ID, &b.URL, &shortDesc, &longDesc, &tags, &researchLevel); err != nil {
			report.Errors = append(report.Errors, "scan error: "+err.Error())
			continue
		}

		if shortDesc.Valid {
			b.Title = &shortDesc.String
		}
		if longDesc.Valid {
			b.Description = &longDesc.String
		}
		if tags.Valid {
			b.Tags = tags.String
		}
		if researchLevel.Valid {
			rl := researchLevel.String
			if rl == "done" {
				rl = "done"
			} else {
				rl = "pending"
			}
			b.ResearchStatus = &rl
		}

		bookmarks = append(bookmarks, b)
	}

	report.Fetched = len(bookmarks)

	tx, err := destDb.Begin()
	if err != nil {
		return nil, fmt.Errorf("tx begin error: %w", err)
	}

	upserted, err := upsertBookmarks(tx, bookmarks)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("upsert error: %w", err)
	}
	report.Upserted = upserted

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx commit error: %w", err)
	}

	return report, nil
}
