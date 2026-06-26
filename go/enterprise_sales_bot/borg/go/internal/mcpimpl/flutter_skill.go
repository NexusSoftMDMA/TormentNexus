package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleRunTest_flutter_skill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	platform, _ :=getString(args, "platform")
	if platform == "" {
		return err("platform is required")
}

	reqBody, _ := json.Marshal(map[string]string{"platform": platform})
	resp, e := http.DefaultClient.Post("https://api.flutterskill.dev/run", "application/json", bytes.NewReader(reqBody))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok("Test started for " + platform)
}

func HandleListPlatforms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	platforms := []string{"Flutter", "React Native", "iOS", "Android", "Web", "Electron", "Tauri", "KMP", ".NET MAUI"}
	data, _ := json.Marshal(platforms)
	return ok(string(data))
}