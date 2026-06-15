package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGcoreDNSZones(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("missing api_key")
}

	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "https://api.gcore.com/dns"
	}
	u, e := url.Parse(base)
	if e != nil {
		return err(e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return err(fmt.Sprintf("gcore api status: %d", resp.StatusCode))
}

	var data interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	return success(fmt.Sprintf("Gcore DNS zones response: %+v", data))
}// touch 1781132126
