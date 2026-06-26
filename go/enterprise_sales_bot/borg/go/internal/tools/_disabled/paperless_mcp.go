package tools

import (
	"context"
	"io"
	"net/http"
)

func HandlePaperlessMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url missing")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request error")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 status")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body error")
}

	return success(string(body))
}// touch 1781132137
