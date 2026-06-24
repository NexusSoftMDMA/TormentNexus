package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchBilibili(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	url := fmt.Sprintf("https://api.bilibili.com/x/web-interface/search/type?search_type=video&keyword=%s", keyword)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
}

	results, found := data["result"].([]interface{})
	if !found || len(results) == 0 {
		return ok("no results found")
}

	first := results[0].(map[string]interface{})
	title := first["title"].(string)
	author := first["author"].(string)
	return ok(fmt.Sprintf("Found video: %s by %s", title, author))
}

func HandleGetBilibiliUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mid, _ :=getInt(args, "mid")
	if mid == 0 {
		return err("mid is required")
}

	url := fmt.Sprintf("https://api.bilibili.com/x/space/acc/info?mid=%d", mid)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
}

	name, _ := data["name"].(string)
	sign, _ := data["sign"].(string)
	return ok(fmt.Sprintf("User: %s - %s", name, sign))
}