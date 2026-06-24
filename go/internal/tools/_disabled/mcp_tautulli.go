package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleTautulliActivity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	key, _ :=getString(args, "api_key")
	url := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_activity", strings.TrimRight(base, "/"), key)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Tautulli: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Activity: %v", result))
}

func HandleTautulliLibraryStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	key, _ :=getString(args, "api_key")
	url := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_library_stats", strings.TrimRight(base, "/"), key)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Tautulli: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Library Stats: %v", result))
}