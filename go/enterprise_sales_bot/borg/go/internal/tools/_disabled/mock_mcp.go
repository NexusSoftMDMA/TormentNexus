package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleStartMock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	port, _ :=getInt(args, "port")
	url, _ :=getString(args, "url")
	if name == "" {
		return err("missing name")
}

	reqURL := fmt.Sprintf("%s/mock/%s?port=%d", url, name, port)
	req, e := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	return ok("mock started: " + name)
}

func HandleStopMock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	url, _ :=getString(args, "url")
	if name == "" {
		return err("missing name")
}

	req, e := http.NewRequestWithContext(ctx, "DELETE", url+"/mock/"+name, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}

	return ok("mock stopped: " + name)
}