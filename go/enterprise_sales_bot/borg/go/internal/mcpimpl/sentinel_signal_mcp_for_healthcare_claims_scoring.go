package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleDenialRisk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	claimID, _ :=getString(args, "claim_id")
	if claimID == "" {
		return err("claim_id is required")
}

	body, _ := json.Marshal(map[string]string{"claim_id": claimID})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.sentinel-signal.com/v1/denial-risk", bytes.NewBuffer(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Denial risk score: %v", result["score"]))
}

func HandlePriorAuth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	patientID, _ :=getString(args, "patient_id")
	if patientID == "" {
		return err("patient_id is required")
}

	body, _ := json.Marshal(map[string]string{"patient_id": patientID})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.sentinel-signal.com/v1/prior-auth", bytes.NewBuffer(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Prior authorization status: %v", result["status"]))
}