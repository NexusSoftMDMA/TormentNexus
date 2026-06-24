package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListGems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gemType, _ :=getString(args, "type")
	url := "https://api.gemsuite.io/gems"
	if gemType != "" {
		url += "?type=" + gemType
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch gems: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Gems: %v", result))
}

func HandleGetGem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := "https://api.gemsuite.io/gems/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch gem: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Gem: %v", result))
}