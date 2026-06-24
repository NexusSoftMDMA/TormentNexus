package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleWarpGBMTrain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataset, _ :=getString(args, "dataset")
	backend, _ :=getString(args, "backend")
	if dataset == "" {
		return err("dataset argument is required")
	}
	payload := map[string]string{"dataset": dataset, "backend": backend}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.warpgbm.io/train", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("model training started successfully")
}

func HandleWarpGBMPredict(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	artifact, _ :=getString(args, "artifact")
	data, _ :=getString(args, "data")
	if artifact == "" || data == "" {
		return err("artifact and data arguments are required")
	}
	payload := map[string]string{"artifact": artifact, "data": data}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.warpgbm.io/predict", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("prediction request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("prediction completed")
}// touch 1781132144
