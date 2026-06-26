package mcpimpl

import (
	"context"
	"net/url"
)

func HandleDecompose(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	parsed, e := url.Parse(urlStr)
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	result := map[string]interface{}{
		"scheme":   parsed.Scheme,
		"host":     parsed.Host,
		"path":     parsed.Path,
		"query":    parsed.RawQuery,
		"fragment": parsed.Fragment,
	}
	return ok(result)
}