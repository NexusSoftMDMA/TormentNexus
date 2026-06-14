package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListToggles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiUrl, _ :=getString(args, "apiUrl")
	token, _ :=getString(args, "apiToken")
	if apiUrl == "" {
		return err("apiUrl required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", apiUrl+"/api/admin/toggles", nil)
	if e != nil {
		return err("request create: " + e.Error())
}

	req.Header.Set("Authorization", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read: " + e.Error())
}

	var data interface{}
	json.Unmarshal(body, &data)
	return ok(fmt.Sprint(data))
}