package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetChampion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing champion name")
}

	url := fmt.Sprintf("https://raw.githubusercontent.com/Valorant-API/tft-oracle/main/data/champions/%s.json", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Champion %s: %v", name, data))
}

func HandleGetItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing item name")
}

	url := fmt.Sprintf("https://raw.githubusercontent.com/Valorant-API/tft-oracle/main/data/items/%s.json", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Item %s: %v", name, data))
}