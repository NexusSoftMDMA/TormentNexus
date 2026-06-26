package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListAnimations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.lottiefiles.com/v1"
	}
	url := fmt.Sprintf("%s/animations", baseURL)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch animations: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success("animations listed")
}

func HandleGetAnimation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing 'id' argument")
}

	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.lottiefiles.com/v1"
	}
	url := fmt.Sprintf("%s/animations/%s", baseURL, id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch animation: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success("animation details retrieved")
}