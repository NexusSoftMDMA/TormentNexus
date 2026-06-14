package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleKaggleDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	username, _ :=getString(args, "username")
	key, _ :=getString(args, "key")
	if username == "" || key == "" {
		return err("username and key are required")
}

	url := "https://www.kaggle.com/api/v1/datasets/list"
	if query != "" {
		url += "?search=" + query
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.SetBasicAuth(username, key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to make request")
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response")
}

	result, _ := json.MarshalIndent(data, "", "  ")
	return ok("Datasets: " + string(result))
}

func HandleKaggleCompetitions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	key, _ :=getString(args, "key")
	if username == "" || key == "" {
		return err("username and key are required")
}

	url := "https://www.kaggle.com/api/v1/competitions/list"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.SetBasicAuth(username, key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to make request")
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response")
}

	result, _ := json.MarshalIndent(data, "", "  ")
	return ok("Competitions: " + string(result))
}