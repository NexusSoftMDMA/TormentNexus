package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleNuclearData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.nuclear.example.com/data?country=" + getString(args, "country")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return ok("data retrieved")
}

func HandleListReactors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.nuclear.example.com/reactors?limit=" + getString(args, "limit")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list reactors: " + e.Error())
}

	defer resp.Body.Close()
	var list []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&list); e != nil {
		return err("decode error: " + e.Error())
}

	return success("reactors list")
}