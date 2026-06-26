package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetSleep_oura_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("missing access_token")
}

	start, _ :=getString(args, "start_date")
	end, _ :=getString(args, "end_date")
	if start == "" || end == "" {
		return err("missing start_date or end_date")
}

	url := "https://api.ouraring.com/v2/usercollection/sleep?start_date=" + start + "&end_date=" + end
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var data interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to decode: " + e.Error())
}

	jsonBytes, e := json.MarshalIndent(data, "", "  ")
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(jsonBytes))
}