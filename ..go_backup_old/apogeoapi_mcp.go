package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetApogeoInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	url := "https://api.apogeoapi.com/info"
	if city != "" {
		url = fmt.Sprintf("%s?city=%s", url, city)

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
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Apogeo Info: %v", data))
}

}

func HandleGetApogeoStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.apogeoapi.com/status")
	if e != nil {
		return err("status request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read status: " + e.Error())
}

	return ok(fmt.Sprintf("Status: %s", string(body)))
}