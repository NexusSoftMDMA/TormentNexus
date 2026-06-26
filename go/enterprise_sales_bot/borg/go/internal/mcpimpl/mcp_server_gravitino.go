package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListCatalogs_mcp_server_gravitino(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8090"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url+"/api/catalogs", nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	return success("catalogs retrieved")
}

func HandleGetCatalog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8090"
	}
	name, _ :=getString(args, "name")
	if name == "" {
		return err("catalog name required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url+"/api/catalogs/"+name, nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	return success("catalog retrieved")
}// touch 1781132132
