package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleAflTeamInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	team, _ :=getString(args, "team")
	if team == "" {
		return err("team parameter required")
}

	url := "https://api.squiggle.com.au/?q=teams&team=" + team
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return success(string(body))
}