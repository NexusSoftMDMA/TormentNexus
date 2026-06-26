package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleFleetHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	site, _ :=getString(args, "site")
	resp, e := http.DefaultClient.Get("https://mock-unifi.api/fleet?site=" + site)
	if e != nil {
		return err("failed to fetch fleet health: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("json decode error: " + e.Error())
}

	return success("fleet health data retrieved")
}

func HandleAnomalyDetection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	site, _ :=getString(args, "site")
	threshold, _ :=getInt(args, "threshold")
	resp, e := http.DefaultClient.Get("https://mock-unifi.api/anomalies?site=" + site + "&threshold=" + string(rune(threshold)))
	if e != nil {
		return err("failed to fetch anomalies: " + e.Error())
}

	defer resp.Body.Close()
	var data []map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("json decode error: " + e.Error())
}

	return success("anomaly detection complete")
}