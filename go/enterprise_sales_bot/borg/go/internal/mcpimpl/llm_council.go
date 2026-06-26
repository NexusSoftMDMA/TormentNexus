package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCouncilMembers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://llm-council.example.com/members")
	if e != nil {
		return err("failed to fetch members: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var members []map[string]interface{}
	if e := json.Unmarshal(body, &members); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("Council members: " + string(body))
}