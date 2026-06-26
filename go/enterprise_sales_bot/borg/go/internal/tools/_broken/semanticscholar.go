package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchPapers_semanticscholar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.semanticscholar.org/graph/v1/paper/search?query=%s&limit=%d", url.QueryEscape(query), limit)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result struct {
		Data []struct {
			PaperId string `json:"paperId"`
			Title   string `json:"title"`
		} `json:"data"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	if len(result.Data) == 0 {
		return success("No papers found")
}

	msg := "Papers:\n"
	for _, p := range result.Data {
		msg += fmt.Sprintf("- %s (ID: %s)\n", p.Title, p.PaperId)

	return ok(msg)
}

}

func HandleGetPaperDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperId, _ :=getString(args, "paperId")
	if paperId == "" {
		return err("paperId parameter required")
}

	u := fmt.Sprintf("https://api.semanticscholar.org/graph/v1/paper/%s", url.PathEscape(paperId))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var details struct {
		PaperId string `json:"paperId"`
		Title   string `json:"title"`
		Year    int    `json:"year"`
	}
	if e = json.Unmarshal(body, &details); e != nil {
		return err(e.Error())
}

	msg := fmt.Sprintf("Paper: %s\nYear: %d\nID: %s", details.Title, details.Year, details.PaperId)
	return ok(msg)
}