package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetCameraStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	cameraID, _ :=getString(args, "camera_id")
	if apiKey == "" || cameraID == "" {
		return err("missing required parameters: api_key, camera_id")
}

	url := "https://api.rhombus.com/v1/cameras/status?apiKey=" + apiKey + "&cameraId=" + cameraID
	resp, e := http.DefaultClient.Get(url)
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

	return ok(data)
}