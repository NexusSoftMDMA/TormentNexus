package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetWheelSize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	make, _ :=getString(args, "make")
	model, _ :=getString(args, "model")
	if make == "" || model == "" {
		return err("make and model required")
}

	url := "https://api.wheel-size.com/v1/wheels?make=" + make + "&model=" + model
	if year := getString(args, "year"); year != "" {
		url += "&year=" + year
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Wheel size data: %v", data))
}

func HandleGetFitment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	make, _ :=getString(args, "make")
	model, _ :=getString(args, "model")
	if make == "" || model == "" {
		return err("make and model required")
}

	url := "https://api.wheel-size.com/v1/fitment?make=" + make + "&model=" + model
	if year := getString(args, "year"); year != "" {
		url += "&year=" + year
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Fitment data: %v", data))
}