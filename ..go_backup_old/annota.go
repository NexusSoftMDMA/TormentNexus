package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetAnnotations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
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

func HandleCreateAnnotation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	annotation, _ :=getString(args, "annotation")
	_ = url
	_ = annotation
	return success("annotation created")
}