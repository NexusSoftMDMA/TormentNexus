package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleConvexGetDeployment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "deployment_name")
	if name == "" {
		return err("deployment_name is required")
}

	apiKey := os.Getenv("CONVEX_API_KEY")
	if apiKey == "" {
		return err("CONVEX_API_KEY not set")
}

	base := os.Getenv("CONVEX_API_URL")
	if base == "" {
		base = "https://api.convex.cloud"
	}
	url := fmt.Sprintf("%s/api/deployments/%s", base, name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request creation: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err("API error " + resp.Status + ": " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse: " + e.Error())
}

	return ok(fmt.Sprintf("Deployment info: %+v", result))
}