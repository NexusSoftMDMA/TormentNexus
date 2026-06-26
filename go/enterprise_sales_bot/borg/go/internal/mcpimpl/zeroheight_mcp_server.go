package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleZeroheightGetStyleguide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.zeroheight.com/v1/styleguides/"+id, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return success("Styleguide data: " + string(body))
}