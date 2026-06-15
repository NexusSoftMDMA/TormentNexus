package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLei(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lei, _ :=getString(args, "lei")
	if lei == "" {
		return err("lei is required")
}

	url := fmt.Sprintf("https://api.gleif.org/api/v1/lei-records/%s", lei)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
}

	return ok(string(body))
}

func HandleSearchLei(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	page, _ :=getInt(args, "page")
	if page < 1 {
		page = 1
	}
	size, _ :=getInt(args, "size")
	if size < 1 || size > 100 {
		size = 10
	}
	url := fmt.Sprintf("https://api.gleif.org/api/v1/lei-records?filter[legalName]=%s&page[number]=%d&page[size]=%d", name, page, size)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
}

	return ok(string(body))
}