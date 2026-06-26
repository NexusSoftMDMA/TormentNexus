package mcpimpl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTabs_opentabs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://127.0.0.1:9222/json")
	if e != nil {
		return err("failed to fetch tabs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var tabs []map[string]interface{}
	if e = json.Unmarshal(body, &tabs); e != nil {
		return err("failed to parse tabs: " + e.Error())
}

	data, _ := json.Marshal(tabs)
	return ok(string(data))
}

func HandleCloseTab(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument required")
}

	req, e := http.NewRequest("DELETE", fmt.Sprintf("http://127.0.0.1:9222/json/close/%s", url), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to close tab: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("tab close failed with status %d", resp.StatusCode))
}

	return ok("tab closed")
}