package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleRegisterPrototype(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	apiBase := os.Getenv("VIBESHARING_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.vibesharing.io"
	}
	payload, e := json.Marshal(map[string]string{"name": name, "description": desc})
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiBase+"/prototypes", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(fmt.Sprintf("Prototype registered: %v", result["id"]))
}

func HandleGetFeedback(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	protoID, _ :=getString(args, "prototypeId")
	apiBase := os.Getenv("VIBESHARING_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.vibesharing.io"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", apiBase+"/prototypes/"+protoID+"/feedback", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var feedback interface{}
	if e := json.NewDecoder(resp.Body).Decode(&feedback); e != nil {
		return err("failed to decode feedback: " + e.Error())
}

	return ok(fmt.Sprintf("Feedback: %v", feedback))
}