package tools

import (
	"context"
	"io"
	"net/http"
)

type ToolResponse struct {
	Content interface{} `json:"content"`
	IsError bool        `json:"isError,omitempty"`
}

func getString(args map[string]interface{}, key string) string {
	if v, found := args[key]; found {
		if s, found := v.(string); found {
			return s
		}
	}
	return ""
}

func getInt(args map[string]interface{}, key string) int { return 0 }
func getBool(args map[string]interface{}, key string) bool { return false }
func ok(msg string) (ToolResponse, error) { return ToolResponse{Content: msg}, nil }
func err(msg string) (ToolResponse, error) { return ToolResponse{Content: msg, IsError: true}, nil }
func success(msg string) (ToolResponse, error) { return ToolResponse{Content: msg}, nil }

func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

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