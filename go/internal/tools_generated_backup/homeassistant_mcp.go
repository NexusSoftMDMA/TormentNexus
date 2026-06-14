package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url") + "/api/states/" + getString(args, "entity_id")
	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	state, found := result["state"].(string)
	if !found {
		return err("state field missing")
}

	return ok("State: " + state)
}

func HandleCallService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := fmt.Sprintf("%s/api/services/%s/%s", getString(args, "url"), getString(args, "domain"), getString(args, "service"))
	token, _ :=getString(args, "token")
	payload := map[string]interface{}{"entity_id": getString(args, "entity_id")}
	bodyBytes, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return err("unexpected status: " + resp.Status)
}

	return ok("Service called successfully")
}