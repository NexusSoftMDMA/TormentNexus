package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListAddons(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	resp, e := http.DefaultClient.Get("https://addons.mozilla.org/api/v4/addons/?q=" + filter)
	if e != nil {
		return err("failed to fetch addons: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	return ok(string(data))
}

func HandleGetAddonDetail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addonID, _ :=getString(args, "id")
	if addonID == "" {
		return err("missing id")
}

	resp, e := http.DefaultClient.Get("https://addons.mozilla.org/api/v4/addons/addon/" + addonID + "/")
	if e != nil {
		return err("failed to fetch detail: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	return ok(string(data))
}