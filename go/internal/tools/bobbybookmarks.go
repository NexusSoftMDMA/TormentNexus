//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"github.com/tormentnexushq/tormentnexus-go/internal/sync"
)

// HandleBobbyBookmarksSync triggers a synchronization from a BobbyBookmarks instance.
func HandleBobbyBookmarksSync(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ := getString(args, "base_url", "url")
	if baseURL == "" {
		baseURL = "https://bobbybookmarks.com" // Default public instance or user's self-hosted
	}

	dbPath := "tormentnexus.db"
	perPage := getInt(args, "per_page", "limit")
	if perPage <= 0 {
		perPage = 100
	}

	report, errSync := sync.SyncBobbyBookmarks(ctx, dbPath, baseURL, perPage, true, false)
	if errSync != nil {
		return err(errSync.Error())
	}

	out, _ := json.MarshalIndent(report, "", "  ")
	return ok(string(out))
}
