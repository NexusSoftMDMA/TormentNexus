package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleSearchOntology(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	term, _ :=getString(args, "term")
	if term == "" {
		return err("term is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.ebi.ac.uk/ols/api/search?q="+term+"&format=json", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 status: " + resp.Status)
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result: " + e.Error())
}

	return ok(string(data))
}