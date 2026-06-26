package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListCars(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("missing base_url")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/api/v1/cars", nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Cars: %v", data))
}

func HandleGetCarState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("missing base_url")
}

	carID, _ :=getInt(args, "car_id")
	if carID == 0 {
		return err("missing car_id")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/cars/%d/state", base, carID), nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Car state: %v", data))
}