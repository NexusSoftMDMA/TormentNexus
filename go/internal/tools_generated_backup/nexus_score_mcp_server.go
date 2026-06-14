package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCoverage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	doi, _ :=getString(args, "doi")
	if doi == "" {
		return err("DOI is required")
}

	url := fmt.Sprintf("https://api.crossref.org/works/%s", doi)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("Failed to fetch: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return success("No coverage: DOI not found in Crossref")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read body: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("Failed to parse: %v", e))
}

	message, found := data["message"].(map[string]interface{})
	if !found {
		return success("No coverage: No message in response")
}

	_, found = message["DOI"]
	if !found {
		return success("No coverage: DOI not in message")
}

	score := 1.0
	if _, found := message["title"]; !found {
		score = 0.5
	}
	return ok(fmt.Sprintf("Coverage score: %.1f", score))
}