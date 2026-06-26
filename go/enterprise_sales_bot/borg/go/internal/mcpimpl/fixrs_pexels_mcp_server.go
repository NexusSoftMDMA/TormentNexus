package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSearchPhotos_fixrs_pexels_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	page, _ :=getInt(args, "page")
	if page == 0 {
		page = 1
	}
	perPage, _ :=getInt(args, "per_page")
	if perPage == 0 {
		perPage = 15
	}
	url := fmt.Sprintf("https://api.pexels.com/v1/search?query=%s&page=%d&per_page=%d", query, page, perPage)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", os.Getenv("PEXELS_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	data, _ := json.Marshal(result)
	return success(string(data))
}

func HandleSearchVideos_fixrs_pexels_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	page, _ :=getInt(args, "page")
	if page == 0 {
		page = 1
	}
	perPage, _ :=getInt(args, "per_page")
	if perPage == 0 {
		perPage = 15
	}
	url := fmt.Sprintf("https://api.pexels.com/videos/search?query=%s&page=%d&per_page=%d", query, page, perPage)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", os.Getenv("PEXELS_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	data, _ := json.Marshal(result)
	return success(string(data))
}