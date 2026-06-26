package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandleScanMCP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url argument")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to reach server: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("server returned non-200: " + fmt.Sprintf("%d", resp.StatusCode))
}

	headers := resp.Header
	missing := []string{}
	if headers.Get("Content-Security-Policy") == "" {
		missing = append(missing, "Content-Security-Policy")

	if headers.Get("X-Content-Type-Options") == "" {
		missing = append(missing, "X-Content-Type-Options")

	if headers.Get("X-Frame-Options") == "" {
		missing = append(missing, "X-Frame-Options")

	if len(missing) > 0 {
		return ok("Missing security headers: " + fmt.Sprint(missing))
}

	return ok("All common security headers present")
}
}
}
}