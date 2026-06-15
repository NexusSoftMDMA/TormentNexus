package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetGameState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gameID, _ :=getString(args, "gameId")
	if gameID == "" {
		return err("gameId is required")
}

	url := fmt.Sprintf("https://api.uno-online.com/games/%s", gameID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch game: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}

func HandlePlayCard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gameID, _ :=getString(args, "gameId")
	cardIndex, _ :=getInt(args, "cardIndex")
	color, _ :=getString(args, "color")
	if gameID == "" {
		return err("gameId is required")
}

	payload := map[string]interface{}{"cardIndex": cardIndex}
	if color != "" {
		payload["color"] = color
	}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
}

	url := fmt.Sprintf("https://api.uno-online.com/games/%s/play", gameID)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to play card: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}