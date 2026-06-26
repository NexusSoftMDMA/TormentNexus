package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetVideoInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bvid, _ :=getString(args, "bv")
	if bvid == "" {
		return err("missing bv parameter")
}

	resp, e := http.DefaultClient.Get("https://api.bilibili.com/x/web-interface/view?bvid=" + bvid)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
}

	title, found := data["title"].(string)
	if !found {
		return err("no title in data")
}

	return ok("Video title: " + title)
}