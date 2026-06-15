package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetSite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
	}
	resp, e := http.DefaultClient.Get("https://api.neotomadb.org/v2/data/sites/" + id)
	if e != nil {
		return err("failed to fetch site: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}
	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
	}
	sitename, _ := data["sitename"].(string)
	return ok(fmt.Sprintf("Site: %s", sitename))
}

func HandleGetDataset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
	}
	resp, e := http.DefaultClient.Get("https://api.neotomadb.org/v2/data/datasets/" + id)
	if e != nil {
		return err("failed to fetch dataset: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}
	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
	}
	dsname, _ := data["datasetname"].(string)
	return ok(fmt.Sprintf("Dataset: %s", dsname))
}