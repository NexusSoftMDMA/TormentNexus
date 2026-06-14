package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chartType, _ :=getString(args, "type")
	data, _ :=getString(args, "data")
	options, _ :=getString(args, "options")

	payload := map[string]interface{}{
		"type":    chartType,
		"data":    data,
		"options": options,
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.antv.antgroup.com/chart/generate", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to make request: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	chartURL, found := result["url"].(string)
	if !found {
		chartURL, found = result["imageUrl"].(string)

	if found {
		return ok("Chart generated: " + chartURL)
}

	b, _ := json.Marshal(result)
	return ok(string(b))
}
}