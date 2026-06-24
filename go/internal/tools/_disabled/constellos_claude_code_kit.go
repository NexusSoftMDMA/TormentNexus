package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleGenerateTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content required")
	}
	e := ioutil.WriteFile("/tmp/types.ts", []byte(content), 0644)
	if e != nil {
		return err(fmt.Sprintf("write failed: %v", e))
	}
	return ok("file written")
}

func HandleSaveSubagentState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	state, _ :=getString(args, "state")
	if state == "" {
		return err("state required")
	}
	resp, e := http.DefaultClient.Post("http://localhost/save-state", "text/plain", strings.NewReader(state))
	if e != nil {
		return err(fmt.Sprintf("post failed: %v", e))
	}
	defer resp.Body.Close()
	return success("state saved")
}