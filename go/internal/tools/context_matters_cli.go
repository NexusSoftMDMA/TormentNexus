//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleContextSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	body, _ := json.Marshal(map[string]string{"key": key, "value": value})
	resp, e := http.DefaultClient.Post("http://localhost:8080/context", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to set context: " + e.Error())
}

	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("context set successfully")
}

func HandleContextGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	resp, e := http.DefaultClient.Get("http://localhost:8080/context?key=" + key)
	if e != nil {
		return err("failed to get context: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var data map[string]string
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("context: " + data["value"])
}
