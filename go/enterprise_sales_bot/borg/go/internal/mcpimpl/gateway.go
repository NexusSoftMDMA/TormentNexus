package mcpimpl

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleGateway(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return success(string(body))
}

	return ok(data)
}