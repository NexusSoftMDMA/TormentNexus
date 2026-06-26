package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetGtmStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("invalid JSON: %v", e))
}

	gtmVersion, found := data["gtm_version"].(string)
	if !found {
		gtmVersion = "unknown"
	}
	return ok(fmt.Sprintf("GTM status: version=%s", gtmVersion))
}

func HandleRunEnrichment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityID, _ :=getString(args, "entity_id")
	if entityID == "" {
		return err("entity_id is required")
}

	url := fmt.Sprintf("https://api.signaliz.io/enrich/%s", entityID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("enrichment failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("enrichment returned status %d", resp.StatusCode))
}

	return success("Enrichment queued successfully")
}