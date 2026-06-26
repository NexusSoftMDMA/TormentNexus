package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchDartPackages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	u := fmt.Sprintf("https://pub.dev/api/packages?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	packages, found := result["packages"].([]interface{})
	if !found {
		return ok("[]")
}

	names := make([]string, 0)
	for _, p := range packages {
		pkg, found := p.(map[string]interface{})
		if !found {
			continue
		}
		name, found := pkg["name"].(string)
		if found {
			names = append(names, name)

	}
	return success(fmt.Sprintf("%v", names))
}

}

func HandleGetDartPackageInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	packageName, _ :=getString(args, "packageName")
	if packageName == "" {
		return err("packageName parameter is required")
}

	u := fmt.Sprintf("https://pub.dev/api/packages/%s", url.PathEscape(packageName))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	info, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	return ok(string(info))
}