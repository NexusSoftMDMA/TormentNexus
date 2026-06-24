package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetVehicleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vin, _ :=getString(args, "vin")
	if vin == "" {
		return err("vin is required")
}

	url := "https://api.automobile.example/vehicle?vin=" + vin
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return success("Vehicle info: " + string(body))
}

func HandleControlVehicle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vin, _ :=getString(args, "vin")
	action, _ :=getString(args, "action")
	if vin == "" || action == "" {
		return err("vin and action are required")
}

	url := "https://api.automobile.example/control?vin=" + vin + "&action=" + action
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("control failed with status " + resp.Status)
}

	return ok("Action " + action + " applied to vehicle " + vin)
}