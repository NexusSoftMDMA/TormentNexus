package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
)

func HandleGetAgentConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	url := fmt.Sprintf("https://api.fiber-pay.io/v1/agents/%s/config", agentID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch agent config: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("Agent config retrieved")
}

func HandleExecuteAgentAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	action, _ :=getString(args, "action")
	payload := map[string]string{"agent": agentID, "action": action}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	url := fmt.Sprintf("https://api.fiber-pay.io/v1/actions/execute")
	resp, e := http.DefaultClient.Post(url, "application/json", ioutil.NopCloser(struct{ io.Reader }{body:(body)}()))
	if e != nil {
		return err("action failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok(fmt.Sprintf("Action %s executed for agent %s", action, agentID))
}