package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleSwiftPackageInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("package name required")
	}
	url := "https://api.swiftpackageindex.com/api/v1/packages/" + name
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("package info retrieved")
}

func HandleSwiftSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("search query required")
	}
	url := "https://api.swiftpackageindex.com/api/v1/search?q=" + query
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("search completed")
}// touch 1781132141
