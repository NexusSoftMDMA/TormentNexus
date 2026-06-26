package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleJdl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userId, _ :=getInt(args, "userId")
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/todos/%d", userId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse")
}

	title, found := data["title"].(string)
	if !found {
		return err("title not found")
}

	return ok(fmt.Sprintf("Todo: %s", title))
}