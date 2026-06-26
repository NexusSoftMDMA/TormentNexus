package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "device_id")
	if deviceID == "" {
		return err("device_id is required")
}

	resp, e := http.DefaultClient.Get("http://localhost:8080/data/" + deviceID)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(result)
}