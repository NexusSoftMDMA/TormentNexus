package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"os"
)

func HandleGetTeam_withinfocus_tba_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	teamKey, _ :=getString(args, "team_key")
	if teamKey == "" {
		return err("missing team_key")
}

	apiKey := os.Getenv("TBA_API_KEY")
	if apiKey == "" {
		return err("TBA_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://www.thebluealliance.com/api/v3/team/"+teamKey, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-TBA-Auth-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("bad status: " + string(body))
}

	return ok(string(body))
}