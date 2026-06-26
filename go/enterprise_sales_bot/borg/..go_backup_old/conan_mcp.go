package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func SearchConanPackages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://api.conan.io/v1/packages/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data struct {
		Packages []map[string]interface{} `json:"packages"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	if data.Packages == nil {
		data.Packages = []map[string]interface{}{}
	}
	return ok(fmt.Sprintf("Found %d packages", len(data.Packages)))
}

func GetConanPackageDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	packageName, _ :=getString(args, "package_name")
	version, _ :=getString(args, "version")
	if packageName == "" || version == "" {
		return err("package_name and version are required")
}

	url := fmt.Sprintf("https://api.conan.io/v1/packages/%s/%s", packageName, version)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("unmarshal failed: %v", e))
}

	return success(fmt.Sprintf("Package details: %+v", data))
}