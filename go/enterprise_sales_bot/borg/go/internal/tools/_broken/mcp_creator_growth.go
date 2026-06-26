package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetGrowthData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityID, _ :=getString(args, "entity_id")
	metric, _ :=getString(args, "metric")
	url := fmt.Sprintf("https://api.growth.com/v1/data?entity=%s&metric=%s", entityID, metric)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch growth data")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response")
}

	return ok(fmt.Sprintf("Growth data: %v", result))
}

func HandleCreateGrowthEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityID, _ :=getString(args, "entity_id")
	value, _ :=getInt(args, "value")
	payload := map[string]interface{}{
		"entity_id": entityID,
		"value":     value,
	}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload")
}

	resp, e := http.DefaultClient.Post("https://api.growth.com/v1/entries", "application/json", nil)
	if e != nil {
		return err("failed to create entry")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(fmt.Sprintf("Created entry: %s", string(body)))
}