package main

import (
	"bobbybookmarks/internal/database"
	"bobbybookmarks/internal/models"
	"log"
	"time"
)

func main() {
	database.InitDB()
	log.Println("BobbyBookmarks Research Worker Active...")

	for {
		processNext()
		time.Sleep(5 * time.Second)
	}
}

func processNext() {
	var bm models.Bookmark
	err := database.DB.Get(&bm, "SELECT * FROM bookmarks WHERE research_status = 'pending' LIMIT 1")
	if err != nil {
		return // No pending bookmarks
	}

	log.Printf("Researching: %s\n", bm.URL)

	// Mark as running
	database.DB.Exec("UPDATE bookmarks SET research_status = 'running' WHERE id = ?", bm.ID)

	// TODO: Add real scraping/LLM logic here (ported from Python deep_research.py)
	// For now, simulating work...
	time.Sleep(2 * time.Second)

	// Mark as done
	database.DB.Exec("UPDATE bookmarks SET research_status = 'done', researched_at = ? WHERE id = ?", time.Now(), bm.ID)
	log.Printf("Completed: %s\n", bm.URL)
}
