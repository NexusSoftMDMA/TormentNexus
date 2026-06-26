package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleStartExperiment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	name, _ :=getString(args, "name")
	config, _ :=getString(args, "config")
	reqURL := baseURL + "/experiments/start"
	body := map[string]string{"name": name, "config": config}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request body")
}

	resp, e := http.DefaultClient.Post(reqURL, "application/json", bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("failed to start experiment: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("server returned status " + fmt.Sprintf("%d", resp.StatusCode))
}

	return success("experiment started successfully")
}

func HandleGetExperimentStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	expID, _ :=getString(args, "experiment_id")
	reqURL := baseURL + "/experiments/" + expID + "/status"
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("failed to get status: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("server returned status " + fmt.Sprintf("%d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	status, found := result["status"].(string)
	if !found {
		return err("status field missing")
}

	return ok("Experiment " + expID + " status: " + status)
}