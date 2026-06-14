package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	matchID, _ :=getInt(args, "match_id")
	url := fmt.Sprintf("https://api.opendota.com/api/matches/%d", matchID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch match: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetHeroStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.opendota.com/api/heroStats")
	if e != nil {
		return err("failed to fetch hero stats: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}