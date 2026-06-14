package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetComponentContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "componentId")
	if id == "" {
		return err("componentId required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.usevyre.com/context/"+id, nil)
	if e != nil {
		return err("request error")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("fetch error")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse error")
}

	return ok(fmt.Sprintf("Context: %v", data))
}