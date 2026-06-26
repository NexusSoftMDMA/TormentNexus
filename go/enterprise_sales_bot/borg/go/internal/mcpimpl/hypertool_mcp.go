package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleHypertoolMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	persona, _ :=getString(args, "persona")
	req, e := http.NewRequestWithContext(ctx, "GET", server+"/tools?persona="+url.QueryEscape(persona), nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}