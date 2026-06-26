package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetMatchPrediction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	matchID, _ :=getString(args, "match_id")
	if matchID == "" {
		return err("missing match_id")
}

	url := "https://api.footballbin.com/predictions/" + matchID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}

func HandleGetLeaguePredictions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	league, _ :=getString(args, "league")
	if league == "" {
		return err("missing league")
}

	url := "https://api.footballbin.com/leagues/" + league + "/predictions"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}