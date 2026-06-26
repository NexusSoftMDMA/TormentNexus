package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type narrativeResponse struct {
	Narrative string `json:"narrative"`
}

func HandleGetNarrative(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("missing required parameter: ticker")
}

	url := fmt.Sprintf("https://api.aletaindex.com/v1/narratives/%s", ticker)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch narrative: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var nr narrativeResponse
	if e := json.Unmarshal(body, &nr); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if nr.Narrative == "" {
		return ok("No narrative available for ticker " + ticker)
}

	return success(nr.Narrative)
}