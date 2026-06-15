package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleRender(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	if source == "" {
		return err("source is required")
}

	format, _ :=getString(args, "format")
	if format == "" {
		format = "svg"
	}
	baseURL, _ :=getString(args, "server")
	if baseURL == "" {
		baseURL = "https://www.plantuml.com/plantuml"
	}
	apiURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), format)
	form := url.Values{}
	form.Set("text", source)
	resp, e := http.DefaultClient.PostForm(apiURL, form)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}