package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetByKrs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	krs, _ :=getString(args, "krs")
	if krs == "" {
		return err("KRS number is required")
}

	url := fmt.Sprintf("https://api-krs.ms.gov.pl/api/krs/OdpisPelny/%s", krs)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("Failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Company data: %v", data))
}