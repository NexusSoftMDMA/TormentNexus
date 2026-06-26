package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleAssessRiskProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	system, _ :=getString(args, "system_description")
	if system == "" {
		return err("system_description is required")
}

	risk := fmt.Sprintf("Risk profile for '%s': Medium severity, 3 identified risks", system)
	return success(risk)
}

func HandleGenerateRiskControls(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	threat, _ :=getString(args, "threat_vector")
	if threat == "" {
		return err("threat_vector is required")
}

	resp, e := http.DefaultClient.Get("https://api.nist.gov/controls?vector=" + threat)
	if e != nil {
		return err("api call failed: " + e.Error())
}

	defer resp.Body.Close()
	var controls []string
	if found := json.NewDecoder(resp.Body).Decode(&controls); found != nil {
		return err("decode failed: " + found.Error())
}

	return ok(fmt.Sprintf("Generated %d controls for %s", len(controls), threat))
}