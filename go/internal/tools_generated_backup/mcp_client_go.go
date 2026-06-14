package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed")
}

	return success(string(body))
}

func HandlePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	bodyInput, _ :=getString(args, "body")
	if url == "" {
		return err("url required")
}

	resp, e := http.DefaultClient.Post(url, "application/json", strings.NewReader(bodyInput))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed")
}

	return success(string(body))
}