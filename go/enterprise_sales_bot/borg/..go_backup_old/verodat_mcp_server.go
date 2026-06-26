package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleVerify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, _ :=getString(args, "data")
	if data == "" {
		return err("data is required")
}

	url := "https://api.verodat.com/verify?data=" + url.QueryEscape(data)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("bad status: " + resp.Status)
}

	return success(string(body))
}