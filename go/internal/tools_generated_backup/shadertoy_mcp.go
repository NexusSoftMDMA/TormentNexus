package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetShader(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing shader id")
}

	url := fmt.Sprintf("https://www.shadertoy.com/api/v1/shaders/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("shadertoy api returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	shader, found := result["Shader"].(map[string]interface{})
	if !found {
		return err("no shader data in response")
}

	info, _ := shader["info"].(map[string]interface{})
	name, _ := info["name"].(string)
	desc, _ := info["description"].(string)
	return ok(fmt.Sprintf("Shader: %s\nDescription: %s", name, desc))
}