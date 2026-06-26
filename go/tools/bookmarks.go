package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

type Bookmark struct {
	Title string   `json:"title"`
	URL   string   `json:"url"`
	Tags  []string `json:"tags"`
}

var bookmarksFile = "./.supercli/bookmarks.json"

func (r *Registry) registerBookmarkTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "add_bookmark",
		Description: "Saves a URL with title and tags (Bobbybookmarks parity). Arguments: title (string), url (string), tags (array of strings)",
		Execute: func(args map[string]interface{}) (string, error) {
			title, _ := args["title"].(string)
			url, _ := args["url"].(string)
			tagsInterface, _ := args["tags"].([]interface{})

			var tags []string
			for _, t := range tagsInterface {
				tags = append(tags, t.(string))
			}

			// Load existing
			var bookmarks []Bookmark
			data, _ := os.ReadFile(bookmarksFile)
			json.Unmarshal(data, &bookmarks)

			// Add new
			bookmarks = append(bookmarks, Bookmark{Title: title, URL: url, Tags: tags})

			// Save
			newData, _ := json.MarshalIndent(bookmarks, "", "  ")
			os.MkdirAll("./.supercli", 0755)
			os.WriteFile(bookmarksFile, newData, 0644)

			return fmt.Sprintf("Bookmark '%s' saved successfully.", title), nil
		},
	})
}
