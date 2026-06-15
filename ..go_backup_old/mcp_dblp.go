package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type dblpResponse struct {
	Result struct {
		Hits struct {
			Hit []struct {
				Info struct {
					Title string `json:"title"`
					Year  string `json:"year"`
				} `json:"info"`
			} `json:"hit"`
		} `json:"hits"`
	} `json:"result"`
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	u := fmt.Sprintf("https://dblp.org/search/publ/api?q=%s&format=json&h=5", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var d dblpResponse
	if e := json.Unmarshal(body, &d); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	hits := d.Result.Hits.Hit
	if len(hits) == 0 {
		return ok("No results found")
}

	result := fmt.Sprintf("Found %d results:\n", len(hits))
	for i, h := range hits {
		result += fmt.Sprintf("%d. %s (%s)\n", i+1, h.Info.Title, h.Info.Year)

	return ok(result)
}
}