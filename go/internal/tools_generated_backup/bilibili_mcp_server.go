package tools

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

func HandleGetUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mid, _ :=getInt(args, "mid")
	url := "https://api.bilibili.com/x/space/acc/info?mid=" + strconv.Itoa(int(mid))
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}