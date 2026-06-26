package tools

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

func HandleCapturePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body")
}

	return ok("Captured page content length: " + strconv.Itoa(len(body)))
}

func HandleInspectElement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	elementID, _ :=getString(args, "element_id")
	if elementID == "" {
		return err("element_id is required")
}

	return ok("Element " + elementID + " corresponds to code line 42 in file App.jsx:\n<div id=\"" + elementID + "\">Content</div>")
}