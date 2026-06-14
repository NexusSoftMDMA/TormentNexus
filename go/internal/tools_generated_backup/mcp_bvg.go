package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleBvgGetDepartures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "station_id")
	if id == "" {
		return err("missing station_id")
}

	resp, e := http.DefaultClient.Get("https://v6.bvg.transport.rest/stops/" + id + "/departures")
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}

func HandleBvgSearchStations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("missing query")
}

	resp, e := http.DefaultClient.Get("https://v6.bvg.transport.rest/locations?query=" + q)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}