package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func HandleListTasks_ticktick_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.ticktick.com/api/v2/tasks")
	if e != nil {
		return err("fetch tasks: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
}

	var tasks []map[string]interface{}
	if e := json.Unmarshal(body, &tasks); e != nil {
		return err("parse tasks: " + e.Error())
}

	return ok("found " + strconv.Itoa(len(tasks)) + " tasks")
}