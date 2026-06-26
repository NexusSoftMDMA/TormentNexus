package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGetCitationCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	doi, _ :=getString(args, "doi")
	if doi == "" {
		return err("doi is required")
}

	apiURL := fmt.Sprintf("https://api.crossref.org/works/%s", url.PathEscape(doi))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Message struct {
			IsReferences        bool `json:"is-referenced-by-count"`
			ReferencesCount     int  `json:"references-count"`
			IsReferencedByCount int  `json:"is-referenced-by-count"`
		} `json:"message"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("DOI %s: citations received: %d, references: %d", doi, data.Message.IsReferencedByCount, data.Message.ReferencesCount))
}

func HandleSearchPapers_citation_intelligence(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL := fmt.Sprintf("https://api.crossref.org/works?query=%s&rows=5", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Message struct {
			Items []struct {
				Title []string `json:"title"`
				DOI   string   `json:"DOI"`
			} `json:"items"`
		} `json:"message"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response")
}

	var titles []string
	for _, item := range result.Message.Items {
		title := ""
		if len(item.Title) > 0 {
			title = item.Title[0]
		}
		titles = append(titles, fmt.Sprintf("%s (DOI: %s)", title, item.DOI))

	return success(fmt.Sprintf("Found %d papers:\n%s", len(titles), joinStrings(titles, "\n")))
}

}

func joinStrings_citation_intelligence(strs []string, sep string) string {
	res := ""
	for i, s := range strs {
		if i > 0 {
			res += sep
		}
		res += s
	}
	return res
}