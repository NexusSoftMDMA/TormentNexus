package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	files, e := ioutil.ReadDir(path)
	if e != nil {
		return err(fmt.Sprintf("failed to read directory: %v", e))
}

	var names []string
	for _, f := range files {
		names = append(names, f.Name())

	return ok(strings.Join(names, "\n"))
}
}