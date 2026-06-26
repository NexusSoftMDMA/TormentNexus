package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleProxy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	payload, _ :=getString(args, "payload")
	if target == "" {
		return err("missing target")
	}
	reqBody := strings.NewReader(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", target, reqBody)
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(e.Error())
	}
	return success("proxied")
}

func HandleConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		return err("missing server")
	}
	return success("connected to " + server)
}