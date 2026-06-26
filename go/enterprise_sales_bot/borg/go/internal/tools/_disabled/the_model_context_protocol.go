package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	data := map[string]string{"echo": msg}
	return ok(fmt.Sprintf("Echo: %s", msg), data)
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	data := map[string]int{"result": sum}
	return ok(fmt.Sprintf("Sum: %d", sum), data)
}

func HandleHttpGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required", nil)
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e), nil)
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("JSON decode failed: %v", e), nil)
}

	return success("HTTP GET succeeded", result)
}