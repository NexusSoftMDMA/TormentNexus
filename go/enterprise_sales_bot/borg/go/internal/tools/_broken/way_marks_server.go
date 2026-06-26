package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListWaymarks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "https://api.waymark.com"
	}
	page, _ :=getInt(args, "page")
	url := fmt.Sprintf("%s/waymarks?page=%d", base, page)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Got %d waymarks", int(result["count"].(float64))))
}

func HandleCreateWaymark(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "https://api.waymark.com"
	}
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	payload, e := json.Marshal(map[string]string{"name": name, "description": desc})
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	resp, e := http.DefaultClient.Post(base+"/waymarks", "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Created waymark '%s'", name))
}