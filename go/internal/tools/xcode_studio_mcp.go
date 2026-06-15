//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	switch action {
	case "list":
		resp, e := http.DefaultClient.Get("http://localhost:5173/api/projects")
		if e != nil {
			return err("failed to request: " + e.Error())
}

		defer resp.Body.Close()
		io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			return err("server returned status: " + resp.Status)
		return ok("list successful")
	case "open":
		projectPath, _ :=getString(args, "projectPath")
		if projectPath == "" {
			return err("projectPath is required")
		return ok("opened project: " + projectPath)
	default:
		return err("unknown action: " + action)

}
}
}
}