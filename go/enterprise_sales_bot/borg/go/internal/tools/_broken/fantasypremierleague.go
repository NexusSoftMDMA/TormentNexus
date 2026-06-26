package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const fplBaseURL = "https://fantasy.premierleague.com/api"

type fplElement struct {
	ID            int     `json:"id"`
	FirstName     string  `json:"first_name"`
	SecondName    string  `json:"second_name"`
	Team          int     `json:"team"`
	Photo         string  `json:"photo"`
	TotalPoints   int     `json:"total_points"`
	NowCost       float64 `json:"now_cost"`
	GoalsScored   int     `json:"goals_scored"`
	Assists       int     `json:"assists"`
	ElementType   int     `json:"element_type"`
	WebName       string  `json:"web_name"`
	PointsPerGame string  `json:"points_per_game"`
}

type fplTeam struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code int    `json:"code"`
}

type fplFixture struct {
	ID             int    `json:"id"`
	Event          int    `json:"event"`
	KickoffTime    string `json:"kickoff_time"`
	TeamH          int    `json:"team_h"`
	TeamA          int    `json:"team_a"`
	TeamHScore     *int   `json:"team_h_score"`
	TeamAScore     *int   `json:"team_a_score"`
	TeamHDifficulty int   `json:"team_h_difficulty"`
	TeamADifficulty int   `json:"team_a_difficulty"`
}

type fplBootstrap struct {
	Elements []fplElement `json:"elements"`
	Teams    []fplTeam    `json:"teams"`
}

type fplFixturesResp []fplFixture

func fetchFPL(urlStr string, dest interface{}) error {
	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequest("GET", urlStr, nil)
	if reqErr != nil {
		return reqErr
	}
	req.Header.Set("User-Agent", "FPL-MCP-Go/1.0")
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return httpErr
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("FPL API returned status %d", resp.StatusCode)
}

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(dest)
}

func teamNameMap(teams []fplTeam) map[int]string {
	m := make(map[int]string)
	for _, t := range teams {
		m[t.ID] = t.Name
	}
	return m
}

func HandleGetPlayers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var bootstrap fplBootstrap
	fetchErr := fetchFPL(fplBaseURL+"/bootstrap-static/", &bootstrap)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	teamMap := teamNameMap(bootstrap.Teams)

	var filterTeam string
	if v, found := args["team"]; found {
		filterTeam = fmt.Sprintf("%v", v)

	var lines []string
	for _, p := range bootstrap.Elements {
		tName := teamMap[p.Team]
		if filterTeam != "" && tName != filterTeam {
			continue
		}
		price := p.NowCost / 10.0
		line := fmt.Sprintf("%s (%s) | Team: %s | Points: %d | Price: £%.1fm | Goals: %d | Assists: %d",
			p.WebName, p.SecondName, tName, p.TotalPoints, price, p.GoalsScored, p.Assists)
		lines = append(lines, line)

	result := "Fantasy Premier League Players:\n\n"
	for _, l := range lines {
		result += l + "\n"
	}
	result += fmt.Sprintf("\nTotal: %d players", len(lines))

	return ok(result)
}

}
}

func HandleGetFixtures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var bootstrap fplBootstrap
	bootErr := fetchFPL(fplBaseURL+"/bootstrap-static/", &bootstrap)
	if bootErr != nil {
		return err(bootErr.Error())
}

	teamMap := teamNameMap(bootstrap.Teams)

	var fixtures fplFixturesResp
	fixErr := fetchFPL(fplBaseURL+"/fixtures/", &fixtures)
	if fixErr != nil {
		return err(fixErr.Error())
}

	var filterEvent string
	if v, found := args["gameweek"]; found {
		filterEvent = fmt.Sprintf("%v", v)

	var lines []string
	for _, f := range fixtures {
		gw := strconv.Itoa(f.Event)
		if filterEvent != "" && gw != filterEvent {
			continue
		}
		homeName := teamMap[f.TeamH]
		awayName := teamMap[f.TeamA]
		scoreLine := "vs"
		if f.TeamHScore != nil && f.TeamAScore != nil {
			scoreLine = fmt.Sprintf("%d - %d", *f.TeamHScore, *f.TeamAScore)

		kickoff := f.KickoffTime
		if len(kickoff) > 19 {
			kickoff = kickoff[:19]
		}
		line := fmt.Sprintf("GW%d: %s %s %s | Kickoff: %s | Difficulty: H%d/A%d",
			f.Event, homeName, scoreLine, awayName, kickoff, f.TeamHDifficulty, f.TeamADifficulty)
		lines = append(lines, line)

	result := "Fantasy Premier League Fixtures:\n\n"
	for _, l := range lines {
		result += l + "\n"
	}
	result += fmt.Sprintf("\nTotal: %d fixtures", len(lines))

	return ok(result)
}
}
}
}