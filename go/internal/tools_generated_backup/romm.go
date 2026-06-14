package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListGames(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:4200"
	}
	url := fmt.Sprintf("%s/api/games", baseURL)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("games: %v", result))
}