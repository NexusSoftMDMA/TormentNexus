package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCurrentSwissTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://worldtimeapi.org/api/timezone/Europe/Zurich")
	if e != nil {
		return err("failed to fetch time: " + e.Error())
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

	datetime, found := result["datetime"].(string)
	if !found {
		return err("datetime not found")
}

	return ok("Current Swiss time: " + datetime)
}

func HandleSwissGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Grüezi! Welcome to MCP Swiss.")
}