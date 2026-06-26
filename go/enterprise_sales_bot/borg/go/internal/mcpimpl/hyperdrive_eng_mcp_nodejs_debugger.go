package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleAttach(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	protocol, _ :=getString(args, "protocol")
	if host == "" || port == 0 {
		return err("host and port required")
	}
	scheme := "http"
	if protocol != "" {
		scheme = protocol
	}
	urlStr := fmt.Sprintf("%s://%s:%d/json/version", scheme, host, port)
	resp, e := http.DefaultClient.Get(urlStr)
	if e != nil {
		return err(fmt.Sprintf("Failed to connect: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Unexpected status: %d", resp.StatusCode))
	}
	return ok("Attached to Node.js debugger")
}

func HandleEvaluate_hyperdrive_eng_mcp_nodejs_debugger(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	expression, _ :=getString(args, "expression")
	sessionId, _ :=getString(args, "sessionId")
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	if expression == "" || sessionId == "" || host == "" || port == 0 {
		return err("expression, sessionId, host, port required")
	}
	urlStr := fmt.Sprintf("http://%s:%d/json/evaluate", host, port)
	body := map[string]interface{}{
		"expression": expression,
		"sessionId":  sessionId,
	}
	jsonData, _ := json.Marshal(body)
	resp, e := http.DefaultClient.Post(urlStr, "application/json", bytes.NewReader(jsonData))
	if e != nil {
		return err(fmt.Sprintf("Evaluation failed: %v", e))
	}
	defer resp.Body.Close()
	return success("Evaluation sent")
}