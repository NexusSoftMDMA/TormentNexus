package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetSaju(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	birthdate, _ :=getString(args, "birthdate")
	url := "https://api.sazu.app/saju?name=" + name + "&birthdate=" + birthdate
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("SAZU API call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetManse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	birthdate, _ :=getString(args, "birthdate")
	url := "https://api.sazu.app/manse?name=" + name + "&birthdate=" + birthdate
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("SAZU API call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}