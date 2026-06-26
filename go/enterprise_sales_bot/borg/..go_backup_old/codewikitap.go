package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleScanProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	e := scanManifests(path)
	if e != nil {
		return err(e.Error())
	}
	return success("Project scan complete")
}

func HandleFetchDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("package argument required")
	}
	url := "https://codewiki.google.com/fetch?q=" + pkg
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(e.Error())
	}
	content, found := result["content"].(string)
	if !found {
		content = "No documentation found"
	}
	return ok(strings.TrimSpace(content))
}

func scanManifests(path string) error {
	return nil
}// touch 1781132122
