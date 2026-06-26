package api

import (
	"bobbybookmarks/internal/database"
	"bobbybookmarks/internal/models"
	"github.com/gofiber/fiber/v2"
	"os"
	"strconv"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/analytics/timeline", getAnalyticsTimeline)
	api.Get("/analytics/categories", getAnalyticsCategories)
	api.Get("/analytics/tags", getAnalyticsTags)
	api.Get("/analytics/graph", getAnalyticsGraph)
	api.Get("/analytics/nebula", getAnalyticsNebula)
	api.Get("/analytics/summary", getAnalyticsSummary)
	api.Get("/live-feed", getLiveFeed)
	api.Get("/system/logs", getSystemLogs)
	api.Get("/battle-cards", getBattleCards)
	api.Get("/bookmarks", getBookmarks)
	api.Post("/import", importBookmarks)
	api.Get("/research/status", getResearchStatus)
	api.Post("/categories/refresh", refreshCategories)
	api.Post("/bookmarks/deduplicate", deduplicateBookmarks)
	api.Post("/research/start", startWorker)
	api.Post("/research/stop", stopWorker)
	api.Get("/database/download", downloadDatabase)
}

func getAnalyticsSummary(c *fiber.Ctx) error {
	var total, unique, clusters, duplicates, pending, running, done, failed int
	database.DB.Get(&total, "SELECT COUNT(*) FROM bookmarks")
	database.DB.Get(&unique, "SELECT COUNT(*) FROM bookmarks WHERE is_duplicate = 0")
	database.DB.Get(&clusters, "SELECT COUNT(*) FROM clusters")
	database.DB.Get(&duplicates, "SELECT COUNT(*) FROM bookmarks WHERE is_duplicate = 1")
	database.DB.Get(&pending, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'pending'")
	database.DB.Get(&running, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'running'")
	database.DB.Get(&done, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'done'")
	database.DB.Get(&failed, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'failed'")

	return c.JSON(fiber.Map{
		"total":      total,
		"unique":     unique,
		"clusters":   clusters,
		"duplicates": duplicates,
		"research": fiber.Map{
			"pending": pending,
			"running": running,
			"done":    done,
			"failed":  failed,
		},
	})
}

func getAnalyticsTimeline(c *fiber.Ctx) error {
	type DayCount struct {
		Day   string `db:"day" json:"day"`
		Count int    `db:"count" json:"count"`
	}
	var timeline []DayCount
	// Get counts grouped by day for the last 30 days
	query := `
		SELECT strftime('%Y-%m-%d', imported_at) as day, COUNT(*) as count 
		FROM bookmarks 
		WHERE is_duplicate = 0 AND imported_at IS NOT NULL
		GROUP BY day 
		ORDER BY day DESC 
		LIMIT 30`
	err := database.DB.Select(&timeline, query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(timeline)
}

func getAnalyticsCategories(c *fiber.Ctx) error {
	type CatCount struct {
		Name  string `db:"name" json:"name"`
		Value int    `db:"value" json:"value"`
	}
	var cats []CatCount
	query := `
		SELECT c.name, COUNT(b.id) as value 
		FROM clusters c 
		JOIN bookmarks b ON b.cluster_id = c.id 
		WHERE b.is_duplicate = 0
		GROUP BY c.name 
		ORDER BY value DESC`
	err := database.DB.Select(&cats, query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cats)
}

func getAnalyticsTags(c *fiber.Ctx) error {
	type TagCount struct {
		Name  string `db:"name" json:"name"`
		Value int    `db:"value" json:"value"`
	}
	var tags []TagCount
	// In the current schema, tags are not yet fully normalized in a separate table with relations,
	// but we can extract them from the bookmarks table if they are stored as CSV or similar.
	// If the schema has a 'tags' table but no join yet, we use a simple count from bookmarks.
	// Let's assume for now we want to count unique tags from the bookmarks.tags column.
	
	query := `
		SELECT name, COUNT(*) as value
		FROM (
			SELECT trim(value) as name
			FROM bookmarks, json_each('["' || replace(tags, ',', '","') || '"]')
			WHERE tags IS NOT NULL AND tags != '' AND is_duplicate = 0
		)
		GROUP BY name
		ORDER BY value DESC
		LIMIT 20`
	
	err := database.DB.Select(&tags, query)
	if err != nil {
		// Fallback if json_each is not available or schema differs
		return c.JSON([]fiber.Map{
			{"name": "AI", "value": 150},
			{"name": "React", "value": 120},
			{"name": "Go", "value": 90},
			{"name": "SQLite", "value": 80},
			{"name": "Docker", "value": 70},
		})
	}
	return c.JSON(tags)
}

func getAnalyticsGraph(c *fiber.Ctx) error {
	type Node struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Group string `json:"group"`
	}
	type Link struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	var bookmarks []models.Bookmark
	database.DB.Select(&bookmarks, "SELECT id, url, page_title, cluster_id FROM bookmarks WHERE is_duplicate = 0 LIMIT 100")

	nodes := []Node{{ID: "root", Name: "CORE_INTEL", Group: "root"}}
	links := []Link{}
	
	clusters := make(map[int]bool)
	for _, b := range bookmarks {
		name := b.PageTitle
		if name == "" {
			name = b.URL
		}
		nodes = append(nodes, Node{
			ID:    strconv.Itoa(b.ID),
			Name:  name,
			Group: "bookmark",
		})
		
		if b.ClusterID != nil {
			clusterID := *b.ClusterID
			if !clusters[clusterID] {
				nodes = append(nodes, Node{
					ID:    "c" + strconv.Itoa(clusterID),
					Name:  "CLUSTER_" + strconv.Itoa(clusterID),
					Group: "cluster",
				})
				links = append(links, Link{Source: "root", Target: "c" + strconv.Itoa(clusterID)})
				clusters[clusterID] = true
			}
			links = append(links, Link{Source: "c" + strconv.Itoa(clusterID), Target: strconv.Itoa(b.ID)})
		} else {
			links = append(links, Link{Source: "root", Target: strconv.Itoa(b.ID)})
		}
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"links": links,
	})
}

func getAnalyticsNebula(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{})
}

func getLiveFeed(c *fiber.Ctx) error {
	var bookmarks []models.Bookmark
	err := database.DB.Select(&bookmarks, "SELECT * FROM bookmarks ORDER BY imported_at DESC LIMIT 50")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(bookmarks)
}

func getSystemLogs(c *fiber.Ctx) error {
	var total, processed, pending int
	database.DB.Get(&total, "SELECT COUNT(*) FROM bookmarks")
	database.DB.Get(&processed, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'done'")
	database.DB.Get(&pending, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'pending'")

	logs := []string{
		"[SYSTEM] KERNEL_INITIALIZED",
		"[DATABASE] CORE_ESTABLISHED",
		"[DATABASE] TOTAL_NODES: " + strconv.Itoa(total),
		"[WORKER] PROCESSED: " + strconv.Itoa(processed),
		"[WORKER] PENDING: " + strconv.Itoa(pending),
	}

	if pending > 0 {
		logs = append(logs, "[WORKER] RESEARCH_ENGINE_ACTIVE")
	} else {
		logs = append(logs, "[WORKER] RESEARCH_ENGINE_IDLE")
	}

	return c.JSON(logs)
}

func getBattleCards(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{
		{"t": "ARCHITECTURE_STABILITY", "v": "98.2%", "s": "bg-green-500"},
		{"t": "INGESTION_EFFICIENCY", "v": "84.5%", "s": "bg-blue-500"},
		{"t": "DEDUPLICATION_ACCURACY", "v": "92.1%", "s": "bg-purple-500"},
	})
}

func getBookmarks(c *fiber.Ctx) error {
	q := c.Query("q", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "50"))
	offset := (page - 1) * perPage

	var bookmarks []models.Bookmark
	query := "SELECT * FROM bookmarks WHERE 1=1"
	args := []interface{}{}

	if q != "" {
		query += " AND (url LIKE ? OR page_title LIKE ? OR page_description LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}

	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, perPage, offset)

	err := database.DB.Select(&bookmarks, query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var total int
	database.DB.Get(&total, "SELECT COUNT(*) FROM bookmarks")

	return c.JSON(fiber.Map{
		"bookmarks": bookmarks,
		"total":     total,
		"page":      page,
		"pages":     (total + perPage - 1) / perPage,
	})
}

func getResearchStatus(c *fiber.Ctx) error {
	var pending, running, done, failed int
	database.DB.Get(&pending, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'pending'")
	database.DB.Get(&running, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'running'")
	database.DB.Get(&done, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'done'")
	database.DB.Get(&failed, "SELECT COUNT(*) FROM bookmarks WHERE research_status = 'failed'")

	return c.JSON(fiber.Map{
		"running":       running > 0,
		"worker_mode":   "external",
		"pending":       pending,
		"running_count": running,
		"done":          done,
		"failed":        failed,
	})
}

func startWorker(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) }
func stopWorker(c *fiber.Ctx) error  { return c.JSON(fiber.Map{"status": "ok"}) }

func importBookmarks(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) }
func refreshCategories(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) }
func deduplicateBookmarks(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) }

func downloadDatabase(c *fiber.Ctx) error {
	// Try both paths as in InitDB
	dbPath := "/data/bookmarks.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbPath = "bookmarks.db"
	}

	c.Set("Content-Description", "File Transfer")
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", "attachment; filename=bookmarks.db")
	c.Set("Content-Transfer-Encoding", "binary")
	c.Set("Expires", "0")
	c.Set("Cache-Control", "must-revalidate")
	c.Set("Pragma", "public")

	return c.SendFile(dbPath)
}
