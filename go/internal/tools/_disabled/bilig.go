package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uid, _ :=getString(args, "uid")
	if uid == "" {
		return err("uid is required")
}

	url := fmt.Sprintf("https://api.bilibili.com/x/space/acc/info?mid=%s", uid)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Code int `json:"code"`
		Data struct {
			Name  string `json:"name"`
			Level int    `json:"level"`
		} `json:"data"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	if result.Code != 0 {
		return err(fmt.Sprintf("API error code %d", result.Code))
}

	msg := fmt.Sprintf("User: %s, Level: %d", result.Data.Name, result.Data.Level)
	return ok(msg)
}