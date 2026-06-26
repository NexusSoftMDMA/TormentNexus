package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func HandleGetWage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workerID, _ :=getString(args, "worker_id")
	if workerID == "" {
		return err("worker_id is required")
}

	url := fmt.Sprintf("https://api.swarmwage.dev/worker/%s", workerID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Wage int `json:"wage"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Wage: %d", result.Wage))
}

func HandleCalculateWage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hours, _ :=getInt(args, "hours")
	rate, _ :=getInt(args, "rate")
	total := hours * rate
	return success(fmt.Sprintf("Total wage: %d (hours: %d, rate: %d)", total, hours, rate))
}