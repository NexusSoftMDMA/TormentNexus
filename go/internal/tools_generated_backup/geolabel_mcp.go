package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleGeolabel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	latS, _ :=getString(args, "lat")
	lngS, _ :=getString(args, "lng")
	if latS == "" || lngS == "" {
		return err("lat and lng are required")
}

	lat, e := strconv.ParseFloat(latS, 64)
	if e != nil {
		return err("invalid lat")
}

	lng, e := strconv.ParseFloat(lngS, 64)
	if e != nil {
		return err("invalid lng")
}

	u := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f", lat, lng)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("User-Agent", "GeolabelMcp/1.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response")
}

	label, found := data["display_name"].(string)
	if !found {
		return err("no label found")
}

	return ok(label)
}