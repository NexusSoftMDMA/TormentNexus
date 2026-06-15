package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.genlobe.com/docs/endpoints"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("Endpoints: %v", result))
}

func HandleGetEndpointDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		return err("missing required argument: endpoint")
}

	url := fmt.Sprintf("https://api.genlobe.com/docs/endpoints/%s", endpoint)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("Details: %v", result))
}