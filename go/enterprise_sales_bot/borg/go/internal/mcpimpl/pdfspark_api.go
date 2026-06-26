package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleConvertPdf(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	reqURL := fmt.Sprintf("https://api.pdfspark.com/convert?url=%s", url)
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok("Conversion successful")
}