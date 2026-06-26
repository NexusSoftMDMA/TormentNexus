package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

type ToolResponse struct {
	Content interface{} `json:"content"`
	IsError bool        `json:"isError,omitempty"`
}

			return s
		}
	}
	return ""
}

func ok(msg string) (ToolResponse, error) { return ToolResponse{Content: msg}, nil }

func HandleScrape_webscraping_ai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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