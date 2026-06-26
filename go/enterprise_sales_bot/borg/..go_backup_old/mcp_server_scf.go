package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListFrameworks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiBase, _ :=getString(args, "api_base")
	if apiBase == "" {
		return err("api_base is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", apiBase+"/frameworks", nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}

func HandleGetControl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiBase, _ :=getString(args, "api_base")
	controlID, _ :=getString(args, "control_id")
	if apiBase == "" || controlID == "" {
		return err("api_base and control_id are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", apiBase+"/controls/"+controlID, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Control: %v", result))
}