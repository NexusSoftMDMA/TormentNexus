package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetEphemeris(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	body, _ :=getString(args, "body")
	target, _ :=getString(args, "target")
	date, _ :=getString(args, "date")
	lat, _ :=getString(args, "latitude")
	lon, _ :=getString(args, "longitude")
	url := fmt.Sprintf("https://api.openephemeris.com/v1/compute?body=%s&target=%s&date=%s&lat=%s&lon=%s", body, target, date, lat, lon)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get ephemeris: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("failed to parse: " + e.Error())
}

	raw, _ := json.Marshal(result)
	return ok("ephemeris result: " + string(raw))
}