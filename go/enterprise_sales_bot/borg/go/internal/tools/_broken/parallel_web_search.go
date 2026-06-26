package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func HandleParallelWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	enginesVal, _ :=getString(args, "engines")
	if enginesVal == "" {
		enginesVal = "google,bing,duckduckgo"
	}

	engines := strings.Split(enginesVal, ",")
	maxResults, _ :=getInt(args, "max_results")
	if maxResults <= 0 {
		maxResults = 5
	}

	type searchResult struct {
		Engine  string `json:"engine"`
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	}

	var allResults []searchResult
	client := http.DefaultClient

	for _, engine := range engines {
		engine = strings.TrimSpace(strings.ToLower(engine))
		switch engine {
		case "google":
			searchURL := "https://www.google.com/search?q=" + url.QueryEscape(query) + "&num=" + strconv.Itoa(maxResults)
			req, reqErr := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
			if reqErr != nil {
				continue
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ParallelWebSearch/1.0)")
			resp, fetchErr := client.Do(req)
			if fetchErr != nil {
				continue
			}
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				continue
			}
			bodyStr := string(body)
			results := parseGoogleResults(bodyStr, maxResults)
			for i := range results {
				results[i].Engine = "google"
			}
			allResults = append(allResults, results...)

		case "bing":
			searchURL := "https://www.bing.com/search?q=" + url.QueryEscape(query) + "&count=" + strconv.Itoa(maxResults)
			req, reqErr := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
			if reqErr != nil {
				continue
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ParallelWebSearch/1.0)")
			resp, fetchErr := client.Do(req)
			if fetchErr != nil {
				continue
			}
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				continue
			}
			bodyStr := string(body)
			results := parseBingResults(bodyStr, maxResults)
			for i := range results {
				results[i].Engine = "bing"
			}
			allResults = append(allResults, results...)

		case "duckduckgo":
			searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)
			req, reqErr := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
			if reqErr != nil {
				continue
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ParallelWebSearch/1.0)")
			resp, fetchErr := client.Do(req)
			if fetchErr != nil {
				continue
			}
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				continue
			}
			bodyStr := string(body)
			results := parseDuckDuckGoResults(bodyStr, maxResults)
			for i := range results {
				results[i].Engine = "duckduckgo"
			}
			allResults = append(allResults, results...)

	}

	if len(allResults) == 0 {
		return ok("No results found for query: " + query)
}

	resultBytes, jsonErr := json.MarshalIndent(allResults, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(resultBytes))
}

}

func HandleGetSearchEngines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	engines := []map[string]string{
		{"id": "google", "name": "Google", "description": "Search the web using Google"},
		{"id": "bing", "name": "Bing", "description": "Search the web using Bing"},
		{"id": "duckduckgo", "name": "DuckDuckGo", "description": "Search the web using DuckDuckGo"},
	}
	resultBytes, jsonErr := json.MarshalIndent(engines, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(resultBytes))
}