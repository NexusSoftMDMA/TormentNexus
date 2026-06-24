package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func HandleSerpapiSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "q")
	if query == "" {
		return err("query parameter 'q' is required")
}

	num, _ :=getInt(args, "num")
	if num <= 0 {
		num = 10
	}
	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		return err("SERPAPI_API_KEY environment variable not set")
}

	u := fmt.Sprintf("https://serpapi.com/search?q=%s&num=%d&api_key=%s", url.QueryEscape(query), num, apiKey)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to call SerpAPI: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("SerpAPI returned status " + strconv.Itoa(resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	results, found := data["organic_results"].([]interface{})
	if !found {
		return err("no organic_results in response")
}

	return success(fmt.Sprintf("Found %d results for %q", len(results), query))
}