package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLeagues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.openligadb.de/v1/getavailableleagues"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch leagues: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var leagues []struct {
		LeagueName string `json:"leagueName"`
		LeagueID   int    `json:"leagueId"`
	}
	if e := json.Unmarshal(body, &leagues); e != nil {
		return err(fmt.Sprintf("failed to parse leagues: %v", e))
}

	result := ""
	for _, l := range leagues {
		result += fmt.Sprintf("%d: %s\n", l.LeagueID, l.LeagueName)

	return ok(result)
}

}

func HandleGetMatches(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	leagueId, _ :=getString(args, "leagueId")
	if leagueId == "" {
		return err("leagueId parameter is required")
}

	url := fmt.Sprintf("https://api.openligadb.de/v1/getmatchdata/%s", leagueId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch matches: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var matches []struct {
		MatchDateTime string `json:"matchDateTime"`
		Team1         string `json:"team1"`
		Team2         string `json:"team2"`
	}
	if e := json.Unmarshal(body, &matches); e != nil {
		return err(fmt.Sprintf("failed to parse matches: %v", e))
}

	result := ""
	for _, m := range matches {
		result += fmt.Sprintf("%s: %s vs %s\n", m.MatchDateTime, m.Team1, m.Team2)

	return ok(result)
}
}