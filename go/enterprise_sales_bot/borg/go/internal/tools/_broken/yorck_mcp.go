package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleGetMovies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	movies := []map[string]string{
		{"title": "A Quiet Place: Day One", "id": "aqp1"},
		{"title": "Kinds of Kindness", "id": "kok"},
	}
	data, e := json.Marshal(movies)
	if e != nil {
		return err("failed to encode movies")
}

	return ok(string(data))
}

func HandleGetShowtimes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	movieID, _ :=getString(args, "movie")
	if movieID == "" {
		return err("missing 'movie' argument")
}

	url := fmt.Sprintf("https://www.yorck.de/api/showtimes?movie=%s&date=%s", movieID, getString(args, "date"))
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch showtimes")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}