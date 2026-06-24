package tools

import (
	"context"
	"encoding/json"
)

func HandleListWindows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filterTitle, _ :=getString(args, "title")
	windows := []map[string]interface{}{
		{"id": 1, "title": "Terminal"},
		{"id": 2, "title": "Browser"},
	}
	var filtered []map[string]interface{}
	for _, w := range windows {
		if filterTitle == "" || w["title"] == filterTitle {
			filtered = append(filtered, w)

	}
	data, e := json.Marshal(filtered)
	if e != nil {
		return err("failed to marshal windows")
}

	return ok(string(data))
}

}

func HandleGetActiveWindow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	monitor, _ :=getString(args, "monitor")
	window := map[string]interface{}{"id": 3, "title": "Active Terminal", "monitor": "eDP-1"}
	if monitor != "" && window["monitor"] != monitor {
		return err("no active window on monitor " + monitor)
}

	data, e := json.Marshal(window)
	if e != nil {
		return err("failed to marshal window")
}

	return ok(string(data))
}