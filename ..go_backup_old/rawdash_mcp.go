package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleDashboard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	url := "https://api.rawdash.com/dashboards/" + id
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("request creation failed")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 status")
}

	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed")
}

	return success("dashboard fetched")
}

func HandleWidget(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	widgetID, _ :=getString(args, "widget_id")
	if widgetID == "" {
		return err("missing widget_id")
}

	url := "https://api.rawdash.com/widgets/" + widgetID
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("request creation failed")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 status")
}

	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed")
}

	return success("widget fetched")
}// touch 1781132139
