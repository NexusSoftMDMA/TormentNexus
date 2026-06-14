package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func SearchYoutube(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	key := os.Getenv("YOUTUBE_API_KEY")
	if key == "" {
		return err("YOUTUBE_API_KEY not set")
}

	u := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&key=%s", q, key)
	r, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("API error: " + e.Error())
}

	defer r.Body.Close()
	b, e := io.ReadAll(r.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var m map[string]interface{}
	if e := json.Unmarshal(b, &m); e != nil {
		return err("JSON parse error: " + e.Error())
}

	return success(fmt.Sprintf("Found %v results", len(m["items"].([]interface{}))))
}