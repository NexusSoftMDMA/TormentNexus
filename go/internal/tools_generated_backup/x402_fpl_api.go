package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleFplPlayer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	playerID, _ :=getInt(args, "player_id")
	if playerID == 0 {
		return err("player_id is required")
}

	url := fmt.Sprintf("https://fantasy.premierleague.com/api/element-summary/%d/", playerID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse: " + e.Error())
}

	historyRaw, found := data["history"]
	if !found {
		return err("no history found")
}

	out, _ := json.MarshalIndent(historyRaw, "", "  ")
	return ok(string(out))
}