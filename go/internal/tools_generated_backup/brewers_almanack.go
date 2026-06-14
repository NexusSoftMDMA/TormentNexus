package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetRandomBrew(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.brewersalmanack.com/random")
	if e != nil {
		return err("failed to fetch random brew: " + e.Error())
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

	return ok(fmt.Sprintf("Random brew: %v", data))
}

func HandleSearchBrews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	url := fmt.Sprintf("https://api.brewersalmanack.com/search?name=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search brews: " + e.Error())
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

	return ok(fmt.Sprintf("Search results: %v", data))
}