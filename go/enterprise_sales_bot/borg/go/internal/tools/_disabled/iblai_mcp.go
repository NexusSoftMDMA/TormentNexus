package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetPackageDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("package name is required")
}

	resp, e := http.DefaultClient.Get("https://api.iblai.com/docs/" + pkg)
	if e != nil {
		return err("failed to fetch docs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleListPackages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	packages := []string{"@iblai/core", "@iblai/ui", "@iblai/utils"}
	data, e := json.Marshal(packages)
	if e != nil {
		return err("failed to marshal packages")
}

	return success(string(data))
}