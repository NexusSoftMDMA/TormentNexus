//go:build ignore
// +build ignore

package tools

import (
	"context"
	"net/http"
)

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.aurasdk.com/list"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("list failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("list completed")
}

func HandleGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://api.aurasdk.com/get?id=" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("get failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("got item " + id)
}
