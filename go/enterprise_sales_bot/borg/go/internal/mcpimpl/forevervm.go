package mcpimpl

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleListVms_forevervm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.forevervm.com/vms")
	if e != nil {
		return err("failed to fetch VMs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("bad JSON: " + e.Error())
}

	return ok(string(body))
}

func HandleGetVm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := "https://api.forevervm.com/vms/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get VM: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	return ok(string(body))
}