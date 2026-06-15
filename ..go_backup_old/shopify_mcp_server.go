package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGraphQL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	shopURL, _ :=getString(args, "shop_url")
	accessToken, _ :=getString(args, "access_token")
	query, _ :=getString(args, "query")
	variablesStr, _ :=getString(args, "variables")
	var variables interface{}
	if variablesStr != "" {
		e := json.Unmarshal([]byte(variablesStr), &variables)
		if e != nil {
			return err("invalid variables JSON: " + e.Error())

	}
	payload := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		payload["variables"] = variables
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, shopURL, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Shopify-Access-Token", accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Shopify API error: %d - %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}
}