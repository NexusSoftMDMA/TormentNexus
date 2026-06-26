package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleRunScan_giskard_oasis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	dataset, _ :=getString(args, "dataset")
	if model == "" || dataset == "" {
		return err("model and dataset are required")
}

	url := fmt.Sprintf("http://giskard-oasis.example.com/scan?model=%s&dataset=%s", model, dataset)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to start scan: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("invalid response json: " + e.Error())
}

	return ok(fmt.Sprintf("Scan result: %v", result))
}