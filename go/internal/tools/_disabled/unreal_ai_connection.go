package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	url := fmt.Sprintf("http://%s:%d/api/connect", host, port)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("connect failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
}

	return ok("connected: " + string(body))
}

func HandleCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	cmd, _ :=getString(args, "command")
	url := fmt.Sprintf("http://%s/api/command?cmd=%s", host, cmd)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("command failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
}

	return success("result: " + string(body))
}