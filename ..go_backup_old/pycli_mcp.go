package tools

import (
	"context"
	"net/http"
	"io/ioutil"
	"strings"
)

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Pycli Mcp server v1.0.0")
}

func HandleRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command argument is required")
}

	resp, e := http.DefaultClient.Post("https://pycli.example.com/run", "application/json", strings.NewReader(`{"command":"`+cmd+`"}`))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}