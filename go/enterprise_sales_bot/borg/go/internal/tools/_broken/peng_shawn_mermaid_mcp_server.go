package mcpimpl

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
)

func HandleConvertMermaid(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("mermaid code is required")
}

	theme, _ :=getString(args, "theme")
	u := "https://mermaid.ink/img/" + url.QueryEscape(code)
	if theme != "" {
		u += "?theme=" + url.QueryEscape(theme)

	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	b64 := base64.StdEncoding.EncodeToString(body)
	return success("data:image/png;base64," + b64)
}
}