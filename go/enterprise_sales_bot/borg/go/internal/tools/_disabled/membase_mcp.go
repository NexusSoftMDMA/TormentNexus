package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetKey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("http://localhost:11211/%s", key))
	if e != nil {
		return err("failed to get key: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}

func HandleSetKey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	payload := key + "=" + value
	resp, e := http.DefaultClient.Post("http://localhost:11211/", "text/plain", strings.NewReader(payload))
	if e != nil {
		return err("failed to set key: " + e.Error())
}

	defer resp.Body.Close()
	return success("key set successfully")
}