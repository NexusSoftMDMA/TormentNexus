package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDrops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.pagedrop.com/drops")
	if e != nil {
		return err("failed to fetch drops: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var drops []map[string]interface{}
	if e := json.Unmarshal(body, &drops); e != nil {
		return err("failed to parse drops: " + e.Error())
}

	data, _ := json.Marshal(drops)
	return ok(string(data))
}

func HandleGetDrop_pagedrop_api(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing 'id' parameter")
}

	url := fmt.Sprintf("https://api.pagedrop.com/drops/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch drop: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var drop map[string]interface{}
	if e := json.Unmarshal(body, &drop); e != nil {
		return err("failed to parse drop: " + e.Error())
}

	data, _ := json.Marshal(drop)
	return ok(string(data))
}