package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetVatRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	country, _ :=getString(args, "country")
	if country == "" {
		return err("country parameter required")
}

	url := "https://api.vatapi.com/v1/rate?country_code=" + country
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	rate, found := result["rate"]
	if !found {
		return err("rate not found in response")
}

	return ok(fmt.Sprintf("VAT rate: %v", rate))
}