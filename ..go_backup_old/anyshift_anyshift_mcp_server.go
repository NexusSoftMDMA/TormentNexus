package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListInfrastructure(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resourceType, _ :=getString(args, "type")
	url := fmt.Sprintf("https://api.anyshift.dev/infrastructure/%s", resourceType)
	if resourceType == "" {
		url = "https://api.anyshift.dev/infrastructure"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return ok(string(jsonBytes))
}

func HandleGetInfrastructure(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	resourceType, _ :=getString(args, "type")
	url := fmt.Sprintf("https://api.anyshift.dev/infrastructure/%s/%s", resourceType, id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return ok(string(jsonBytes))
}