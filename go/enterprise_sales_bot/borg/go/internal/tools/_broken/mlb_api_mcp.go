package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type mlbTeam struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type mlbTeamsResponse struct {
	Teams []mlbTeam `json:"teams"`
}

func HandleGetTeams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://statsapi.mlb.com/api/v1/teams", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch teams: " + e.Error())
}

	defer resp.Body.Close()
	var data mlbTeamsResponse
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	var sb strings.Builder
	for _, t := range data.Teams {
		sb.WriteString(fmt.Sprintf("- %s (ID: %d)\n", t.Name, t.ID))

	return ok(sb.String())
}

}

func HandleGetStandings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	leagueID, _ :=getInt(args, "leagueId")
	if leagueID == 0 {
		leagueID = 103
	}
	url := fmt.Sprintf("https://statsapi.mlb.com/api/v1/standings?leagueId=%d", leagueID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch standings: " + e.Error())
}

	defer resp.Body.Close()
	var raw map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&raw); e != nil {
		return err("failed to decode standings: " + e.Error())
}

	records, found := raw["records"].([]interface{})
	if !found {
		return err("no records found")
}

	var sb strings.Builder
	for _, r := range records {
		record, found := r.(map[string]interface{})
		if !found {
			continue
		}
		teamRecords, found := record["teamRecords"].([]interface{})
		if !found {
			continue
		}
		for _, tr := range teamRecords {
			team, found := tr.(map[string]interface{})["team"].(map[string]interface{})
			if !found {
				continue
			}
			name, found := team["name"].(string)
			if !found {
				continue
			}
			teamData, found := tr.(map[string]interface{})
			if !found {
				continue
			}
			wins, _ := teamData["wins"].(float64)
			losses, _ := teamData["losses"].(float64)
			sb.WriteString(fmt.Sprintf("%s: %.0f-%.0f\n", name, wins, losses))

	}
	if sb.Len() == 0 {
		return err("no standings data available")
}

	return ok(sb.String())
}
}