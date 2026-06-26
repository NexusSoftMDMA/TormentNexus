package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleReadPoint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	point, _ :=getString(args, "point")
	url := fmt.Sprintf("https://stier.example.com/api/points/%s", point)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get point: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("Point %s value: %v", point, data["value"]))
}

func HandleGetAlarms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://stier.example.com/api/alarms")
	if e != nil {
		return err("failed to get alarms: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var alarms []interface{}
	if e := json.Unmarshal(body, &alarms); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("Found %d alarms", len(alarms)))
}