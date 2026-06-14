package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func HandleApifyCallActor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := getToken()
	actorID, _ :=getString(args, "actor_id")
	inputStr, _ :=getString(args, "input")
	waitForFinish, _ :=getBool(args, "wait_for_finish")
	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 30
	}
	if actorID == "" {
		return err("actor_id is required")
}

	var input map[string]interface{}
	if inputStr != "" {
		if e := json.Unmarshal([]byte(inputStr), &input); e != nil {
			return err("invalid input JSON")

	}
	body, _ := json.Marshal(map[string]interface{}{
		"actorId": actorID, "input": input, "waitForFinish": waitForFinish, "timeout": timeout,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.apify.com/v2/acts/"+actorID+"/runs", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("missing data in response")
}

	runID, found := data["id"].(string)
	if !found {
		return err("missing run ID")
}

	return ok(fmt.Sprintf("Actor run started with ID: %s", runID))
}

}

func HandleApifyGetDatasetItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := getToken()
	datasetID, _ :=getString(args, "dataset_id")
	format, _ :=getString(args, "format")
	if format == "" {
		format = "json"
	}
	if datasetID == "" {
		return err("dataset_id is required")
}

	url := fmt.Sprintf("https://api.apify.com/v2/datasets/%s/items?format=%s", datasetID, format)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body")
}

	return ok(string(body))
}

func getToken() string {
	if t := os.Getenv("APIFY_TOKEN"); t != "" {
		return t
	}
	return ""
}