package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleRunCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	lang, _ :=getString(args, "language")
	if code == "" {
		return err("code required")
}

	reqBody, e := json.Marshal(map[string]string{"code": code, "language": lang})
	if e != nil {
		return err("marshal error")
}

	resp, e := http.DefaultClient.Post("http://localhost:8080/run", "application/json", strings.NewReader(string(reqBody)))
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error")
}

	return ok(string(body))
}

func HandleListRuns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/runs")
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error")
}

	return ok(string(body))
}