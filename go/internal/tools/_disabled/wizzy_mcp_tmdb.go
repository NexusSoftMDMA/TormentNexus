package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchMovie(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "api_key")
	if query == "" || apiKey == "" {
		return err("query and api_key are required")
}

	url := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s", apiKey, query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	results, found := result["results"].([]interface{})
	if !found {
		return ok("No movies found")
}

	var out string
	for _, item := range results {
		m, found := item.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := m["title"].(string)
		id, _ := m["id"].(float64)
		out += fmt.Sprintf("ID: %.0f - %s\n", id, title)

	return ok(out)
}

}

func HandleGetMovieDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	movieID, _ :=getInt(args, "movie_id")
	apiKey, _ :=getString(args, "api_key")
	if movieID == 0 || apiKey == "" {
		return err("movie_id and api_key are required")
}

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", movieID, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var details map[string]interface{}
	if e = json.Unmarshal(body, &details); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	title, _ := details["title"].(string)
	overview, _ := details["overview"].(string)
	releaseDate, _ := details["release_date"].(string)
	return ok(fmt.Sprintf("Title: %s\nRelease: %s\nOverview: %s", title, releaseDate, overview))
}