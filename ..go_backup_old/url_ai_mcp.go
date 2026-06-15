package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleParseUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	raw, _ :=getString(args, "url")
	parsed, e := url.Parse(raw)
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	result := map[string]string{
		"scheme": parsed.Scheme,
		"host":   parsed.Host,
		"path":   parsed.Path,
		"query":  parsed.RawQuery,
	}
	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleShortenUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get("https://is.gd/create.php?format=simple&url=" + url.QueryEscape(target))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}