package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Agent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func HandleGetAgent_octagon_vc_agents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("https://api.octagon.vc/agents/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch agent")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var agent Agent
	if e := json.Unmarshal(body, &agent); e != nil {
		return err("invalid agent data")
}

	return ok(fmt.Sprintf("Agent: %s (%s)", agent.Name, agent.ID))
}

func HandleListAgents_octagon_vc_agents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.octagon.vc/agents")
	if e != nil {
		return err("failed to list agents")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var agents []Agent
	if e := json.Unmarshal(body, &agents); e != nil {
		return err("invalid agents list")
}

	list := ""
	for _, a := range agents {
		list += fmt.Sprintf("- %s (%s)\n", a.Name, a.ID)

	if list == "" {
		list = "No agents found"
	}
	return ok(list)
}
}