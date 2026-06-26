package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleGetDeputy fetches information about a deputy from the Polish Parliament API.
func HandleGetDeputy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing 'id' argument")
}

	url := fmt.Sprintf("https://api.sejm.gov.pl/sejm/term9/deputies/%s", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %s", e.Error()))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to execute request: %s", e.Error()))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %s", e.Error()))
}

	return ok(fmt.Sprintf("Deputy data: %+v", result))
}