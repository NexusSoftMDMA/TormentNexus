package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleSpaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://huggingface.co/api/spaces?limit=5")
	if e != nil {
		return err("Failed to fetch spaces: " + e.Error())
}

	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read body: " + e.Error())
}

	return success(string(b))
}

func HandleSpace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	space, _ :=getString(args, "space")
	if space == "" {
		return err("Missing required argument: space")
}

	url := "https://huggingface.co/api/spaces/" + space
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch space: " + e.Error())
}

	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read body: " + e.Error())
}

	return success(string(b))
}