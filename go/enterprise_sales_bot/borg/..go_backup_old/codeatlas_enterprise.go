package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListTeams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://codeatlas.example.com/api/teams", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	token, _ :=getString(args, "token")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok("Teams: " + string(data))
}

}

func HandleGetTeamDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	teamID, _ :=getString(args, "teamId")
	if teamID == "" {
		return err("teamId is required")
}

	url := fmt.Sprintf("https://codeatlas.example.com/api/teams/%s", teamID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	token, _ :=getString(args, "token")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok("Team details: " + string(data))
}
}