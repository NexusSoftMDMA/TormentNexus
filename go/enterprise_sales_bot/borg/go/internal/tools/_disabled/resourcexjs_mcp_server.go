package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetResource(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resourceID, _ :=getString(args, "resourceId")
	if resourceID == "" {
		return err("resourceId is required")
}

	url := "https://api.resourcexjs.com/resources/" + resourceID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch resource: %v", e))
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

	return ok(fmt.Sprintf("Resource %s: %v", resourceID, data))
}