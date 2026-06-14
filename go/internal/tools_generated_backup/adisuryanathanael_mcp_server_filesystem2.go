package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(body))
}

func HandleWriteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	content, _ :=getString(args, "content")
	resp, e := http.DefaultClient.Post(url, "text/plain", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to post: %v", e))
}

	defer resp.Body.Close()
	_, e = io.Copy(io.Discard, resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success("written")
}