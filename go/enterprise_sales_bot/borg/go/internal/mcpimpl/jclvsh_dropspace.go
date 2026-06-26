package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleCreateDrop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	content, _ :=getString(args, "content")
	body := fmt.Sprintf(`{"name":"%s","content":"%s"}`, name, content)
	resp, e := http.DefaultClient.Post("https://api.dropspace.space/drops", "application/json", strings.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(data)))
}

	return ok(string(data))
}

func HandleGetDrop_jclvsh_dropspace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://api.dropspace.space/drops/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(data)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	return success(result)
}