package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleEarthquakeQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	minMag, _ :=getInt(args, "minMagnitude")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson&minmagnitude=%d&limit=%d", minMag, limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch earthquake data: " + e.Error())
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

	features, found := result["features"].([]interface{})
	if !found {
		return err("no features found")
}

	return ok(fmt.Sprintf("Found %d earthquakes with min magnitude %d", len(features), minMag))
}