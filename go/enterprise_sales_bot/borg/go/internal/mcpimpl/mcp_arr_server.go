package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListSeries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	key, _ :=getString(args, "api_key")
	url := fmt.Sprintf("%s/api/v3/series?apikey=%s", base, key)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch series: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleListMovies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	key, _ :=getString(args, "api_key")
	url := fmt.Sprintf("%s/api/v3/movie?apikey=%s", base, key)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch movies: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}