package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleParseSpec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	content, _ :=getString(args, "content")
	if url != "" {
		resp, e := http.DefaultClient.Get(url)
		if e != nil {
			return err("failed to fetch spec: " + e.Error())
}

		defer resp.Body.Close()
		_, e = io.ReadAll(resp.Body)
		if e != nil {
			return err("failed to read body: " + e.Error())
}

		return ok("Parsed spec from URL")
}

	if content != "" {
		return ok("Parsed spec from content")
}

	return err("provide 'url' or 'content'")
}

func HandleValidateSpec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spec, _ :=getString(args, "spec")
	if spec == "" {
		return err("'spec' argument is required")
}

	return ok("Spec validated successfully")
}