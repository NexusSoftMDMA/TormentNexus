package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleNewsSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := "https://api.tensorfeed.ai/v1/news/search"
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("query", getString(args, "query"))
	q.Set("category", getString(args, "category"))
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("failed to call news search")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("news search returned status " + fmt.Sprint(resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	return success(fmt.Sprintf("News search results: %v", result))
}

func HandleGetModelData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := "https://api.tensorfeed.ai/v1/models"
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("model", getString(args, "model"))
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("failed to call model data")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("model data returned status " + fmt.Sprint(resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	return success(fmt.Sprintf("Model data: %v", result))
}