package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSearchMovie(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return err("TMDB_API_KEY not set")
}

	url := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s&language=en-US", apiKey, query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandlePopularMovies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return err("TMDB_API_KEY not set")
}

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/popular?api_key=%s&language=en-US", apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}