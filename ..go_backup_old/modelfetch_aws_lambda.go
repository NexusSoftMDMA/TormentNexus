package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleListFunctions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fns, _ :=getString(args, "prefix")
	body, e := listFunctions(fns)
	if e != nil {
		return err("failed to list functions: " + e.Error())
}

	return ok(string(body))
}

func HandleInvokeFunction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "functionName")
	payload, _ :=getString(args, "payload")
	url := os.Getenv("LAMBDA_FUNCTION_URL")
	if url == "" {
		return err("LAMBDA_FUNCTION_URL not set")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(payload))
	if e != nil {
		return err("request error: " + e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Lambda-Function", name)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("invoke error: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}
	return ok(string(body))
}