package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleGetAaveReserves(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subgraphURL, _ :=getString(args, "subgraph_url")
	if subgraphURL == "" {
		subgraphURL = "https://api.thegraph.com/subgraphs/name/aave/protocol-v2"
	}
	query := `{ reserves { id name symbol underlyingAsset decimals } }`
	body, _ := json.Marshal(map[string]string{"query": query})
	req, e := http.NewRequestWithContext(ctx, "POST", subgraphURL, strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("Aave reserves retrieved")
}