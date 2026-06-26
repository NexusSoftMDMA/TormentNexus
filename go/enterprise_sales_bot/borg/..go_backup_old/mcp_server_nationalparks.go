package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetParks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	stateCode, _ :=getString(args, "stateCode")
	if stateCode == "" {
		stateCode = "ALL"
	}
	url := fmt.Sprintf("https://developer.nps.gov/api/v1/parks?stateCode=%s&api_key=YOUR_API_KEY", stateCode)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch parks")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	return success(fmt.Sprintf("Parks: %v", result))
}