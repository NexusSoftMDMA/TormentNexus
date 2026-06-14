package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(fmt.Sprintf("status %d, body len %d", resp.StatusCode, len(body)))
}

func HandlePortCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	if host == "" || port == 0 {
		return err("missing host or port")
}

	addr := fmt.Sprintf("%s:%d", host, port)
	resp, e := http.DefaultClient.Get(fmt.Sprintf("http://%s", addr))
	if e != nil {
		return err(fmt.Sprintf("port check failed: %v", e))
}

	resp.Body.Close()
	return ok(fmt.Sprintf("port %d on %s is open", port, host))
}