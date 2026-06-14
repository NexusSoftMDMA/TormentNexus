package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetIndicators(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	url := fmt.Sprintf("https://webapi.bps.go.id/v1/api/domain?key=%s", apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request indicators: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetStatistic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	domain, _ :=getString(args, "domain")
	variable, _ :=getString(args, "variable")
	url := fmt.Sprintf("https://webapi.bps.go.id/v1/api/data/%s/%s?key=%s", domain, variable, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}