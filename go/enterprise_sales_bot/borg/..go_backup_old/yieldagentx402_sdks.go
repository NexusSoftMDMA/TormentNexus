package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListSdks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/sdks")
	if e != nil {
		return err("failed to fetch sdks: " + e.Error())
}

	defer resp.Body.Close()
	var sdks []string
	if e := json.NewDecoder(resp.Body).Decode(&sdks); e != nil {
		return err("failed to decode: " + e.Error())
}

	return ok(fmt.Sprintf("Available sdks: %v", sdks))
}

func HandleGetSdkInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	url := fmt.Sprintf("https://api.example.com/sdks/%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch sdk info: " + e.Error())
}

	defer resp.Body.Close()
	var info map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&info); e != nil {
		return err("failed to decode: " + e.Error())
}

	return success(fmt.Sprintf("SDK info: %v", info))
}