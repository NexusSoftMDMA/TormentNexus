package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HandleSearchUsLaw searches US legal documents
func HandleSearchUsLaw(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query parameter required")
}

	u := fmt.Sprintf("https://api.usa.gov/legal/search?query=%s", url.QueryEscape(q))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %v results", result["count"]))
}

// HandleGetStatute retrieves a US statute by title and section
func HandleGetStatute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	section, _ :=getString(args, "section")
	if title == "" || section == "" {
		return err("title and section required")
}

	u := fmt.Sprintf("https://api.usa.gov/legal/statute?title=%s&section=%s", url.QueryEscape(title), url.QueryEscape(section))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}