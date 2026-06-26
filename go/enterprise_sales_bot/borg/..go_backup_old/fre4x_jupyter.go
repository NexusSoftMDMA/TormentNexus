package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListKernels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "http://localhost:8888"
	}
	resp, e := http.DefaultClient.Get(base + "/api/kernels")
	if e != nil {
		return err("failed to fetch kernels: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var kernels []interface{}
	if e := json.Unmarshal(body, &kernels); e != nil {
		return err("failed to parse kernels: " + e.Error())
}

	return ok(fmt.Sprintf("Kernels: %d found", len(kernels)))
}

func HandleExecuteCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("execute code placeholder")
}