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

func HandleEsFulltextRetrieve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	maxResults, _ :=getInt(args, "maxResults")
	if maxResults <= 0 {
		maxResults = 10
	}

	baseURL := os.Getenv("AI_MENTORA_API_BASE")
	if baseURL == "" {
		return err("AI_MENTORA_API_BASE not set")
}

	u, e := url.Parse(baseURL + "/es-fulltext-retrieve")
	if e != nil {
		return err(fmt.Sprintf("invalid base URL: %v", e))
}

	q := u.Query()
	q.Set("query", query)
	q.Set("max", strconv.Itoa(maxResults))
	u.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}