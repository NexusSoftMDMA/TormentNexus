package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleVectaraQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	customerId, _ :=getString(args, "customerId")
	query, _ :=getString(args, "query")
	corpusKey, _ :=getString(args, "corpusKey")

	body := map[string]interface{}{
		"query": []map[string]string{{"query": query}},
		"corpusKey": []map[string]string{{"customerId": customerId, "corpusId": corpusKey}},
	}
	reqBytes, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.vectara.io/v1/query", io.NopCloser(strings.NewReader(string(reqBytes))))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return ok("Query result: " + string(respBody))
}