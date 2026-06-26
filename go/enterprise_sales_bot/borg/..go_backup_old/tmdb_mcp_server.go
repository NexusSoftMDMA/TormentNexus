package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleSearchMovie(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return err("TMDB_API_KEY not set")
}

	u := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s", apiKey, url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return success(string(body))
}

func HandleGetMovieDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	movieID, _ :=getInt(args, "movie_id")
	if movieID == 0 {
		return err("movie_id parameter is required")
}

	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return err("TMDB_API_KEY not set")
}

	u := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", movieID, apiKey)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return success(string(body))
}