package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleRunActor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	actorID, _ :=getString(args, "actorId")
	if token == "" || actorID == "" {
		return err("missing token or actorId")
}

	body := strings.NewReader("{}")
	url := fmt.Sprintf("https://api.apify.com/v2/acts/%s/runs?token=%s", actorID, token)
	resp, e := http.DefaultClient.Post(url, "application/json", body)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(fmt.Sprintf("Actor run started: %v", result["data"]))
}

func HandleGetActorRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	runID, _ :=getString(args, "runId")
	if token == "" || runID == "" {
		return err("missing token or runId")
}

	url := fmt.Sprintf("https://api.apify.com/v2/actor-runs/%s?token=%s", runID, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Run status: %v", result["data"]))
}