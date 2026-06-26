package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleCreateRetrospective(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	teamID, _ :=getString(args, "teamId")
	payload, e := json.Marshal(map[string]string{"name": name, "teamId": teamID})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.teamretro.com/retrospectives", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}