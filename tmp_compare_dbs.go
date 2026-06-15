//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	files := []string{
		"db_v1_28413952.db",
		"db_v2_28405760.db",
		"db_v3_28389376.db",
		"db_v4_28368896.db",
		"db_v5_28016640.db",
		"tormentnexus.db",
	}

	tables := []string{
		"mcp_servers", "published_mcp_servers", "tools",
		"imported_sessions", "imported_session_memories", "links_backlog",
	}

	// Header
	for _, f := range files {
		fmt.Printf(",%s", strings.TrimSuffix(f, ".db"))
	}
	fmt.Println()

	for _, t := range tables {
		fmt.Print(t)
		for _, f := range files {
			db, err := sql.Open("sqlite", f)
			if err != nil {
				fmt.Print(",ERR")
				continue
			}
			var cnt int
			err = db.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&cnt)
			if err != nil {
				fmt.Print(",ERR")
			} else {
				fmt.Printf(",%d", cnt)
			}
			db.Close()
		}
		fmt.Println()
	}
}
