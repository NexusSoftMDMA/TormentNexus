//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetAST(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filePath")
	if filePath == "" {
		return err("filePath is required")
	}
	resp, e := http.DefaultClient.Get("http://localhost:8080/ast?path=" + filePath)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	return ok(string(body))
}

func HandleFindSymbol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
	}
	resp, e := http.DefaultClient.Get("http://localhost:8080/find?symbol=" + symbol)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
	}
	return success("found")
}
