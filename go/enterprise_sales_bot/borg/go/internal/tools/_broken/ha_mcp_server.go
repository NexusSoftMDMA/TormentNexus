package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleTurnOnLight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, token := getString(args, "url"), getString(args, "token"), getString(args, "entity_id")
	if url == "" || token == "" || entityID == "" {
		return err("missing required parameter")
}

	body := map[string]interface{}{"entity_id": entityID}
	if b := getInt(args, "brightness"); b > 0 {
		body["brightness"] = b
	}
	payload, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url+"/api/services/light/turn_on", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("Home Assistant returned " + resp.Status)
}

	return ok("light turned on")
}

func HandleActivateScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, token := getString(args, "url"), getString(args, "token"), getString(args, "scene_id")
	if url == "" || token == "" || sceneID == "" {
		return err("missing required parameter")
}

	body := map[string]interface{}{"entity_id": sceneID}
	payload, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url+"/api/services/scene/turn_on", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("Home Assistant returned " + resp.Status)
}

	return ok("scene activated")
}