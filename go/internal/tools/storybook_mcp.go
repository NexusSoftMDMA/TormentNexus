//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListStories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	resp, e := http.DefaultClient.Get(url + "/stories.json")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch stories: %v", e))
}

	defer resp.Body.Close()
	var stories []map[string]interface{	if e := json.NewDecoder(resp.Body).Decode(&stories); e != nil {
		return err(fmt.Sprintf("")
}


-reasoner (deepseek)*
}
}