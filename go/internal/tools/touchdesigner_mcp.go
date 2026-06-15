//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleTouchdesignerExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = os.Getenv("TOUCHDESIGNER_HOST")

	port, _ :=getString(args, "port")
	if port == "" {
		port = os.Getenv("TOUCHDESIGNER_PORT")

	if port == "" {

	}
	op, _ :=getString(args, "op")
	command, _ :=getString(args, "command")
	url := fmt.Sprintf("http://%s:%s/execute?op=%s&command=%s", host, port, op, command)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to execute command: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error,
}


-reasoner (deepseek)*
}
},
}