package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type balldontliePlayersResponse struct {
	Data []balldontliePlayer `json:"data"`
}

type balldontliePlayer struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
	Team      struct {
		FullName string `json:"full_name"`
	} `json:"team"`
}

func HandleGetPlayerStats_nba_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "playerName")
	if name == "" {
		return err("playerName is required")
}

	url := fmt.Sprintf("https://www.balldontlie.io/api/v1/players?search=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch player: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var players balldontliePlayersResponse
	if e := json.Unmarshal(body, &players); e != nil {
		return err("failed to parse response")
}

	if len(players.Data) == 0 {
		return err("no player found")
}

	p := players.Data[0]
	msg := fmt.Sprintf("%s %s - %s (%s)", p.FirstName, p.LastName, p.Position, p.Team.FullName)
	return ok(msg)
}