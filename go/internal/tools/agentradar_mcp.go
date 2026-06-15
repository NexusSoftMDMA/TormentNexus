//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Agent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func HandleSearchAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("https://api.agentradar.com/agents?q=" + query)
	if e != nil {
		return err("failed to fetch agents: " + e.Error())
}

	defer resp.Body.Close()
	var agents []Agent
	if e := json.NewDecoder(resp.Body).Decode(&agents); e != nil {
		return err("failed to decode response: " + e.Error())
}

	result := ""
	for _, a := range agents {
		result += fmt.Sprintf("ID: %s, Name: %s, Role: %s\n", a.ID, a.Name, a.Role)

	return ok(result)
}

}

func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	resp, e := http.DefaultClient.Get("https://api.agentradar.com/agents/" + id)
	if e != nil {
		return err("failed to fetch agent: " + e.Error())
}

	defer resp.Body.Close()
	var agent Agent
	if e := json.NewDecoder(resp.Body).Decode(&agent); e != nil {
		return err("failed to decode agent: " + e.Error())
}

	return ok(fmt.Sprintf("ID: %s, Name: %s, Role: %s", agent.ID, agent.Name, agent.Role))
}
