package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchLaw(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	resp, e := http.DefaultClient.Get("https://api.law.go.kr/search?q=" + query)
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
	return success(fmt.Sprintf("found %v results", result["total"]))
}

func HandleGetLawDetail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lawID, _ :=getString(args, "law_id")
	if lawID == "" {
		return err("law_id is required")
	}
	resp, e := http.DefaultClient.Get("https://api.law.go.kr/detail?id=" + lawID)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
	}
	return success(string(body))
}