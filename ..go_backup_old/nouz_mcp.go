package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleNouz(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Nouz Mcp"
	}

	body, e := json.Marshal(map[string]string{"message": fmt.Sprintf("Hello, %s!", name)})
	if e != nil {
		return err("failed to marshal response")
}

	resp, e := http.DefaultClient.Post("https://httpbin.org/post", "application/json", nil)
	if e != nil {
		return err("failed to post")
}

	defer resp.Body.Close()

	return success(string(body))
}