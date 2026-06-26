package tools

import (
	"context"
	"io/ioutil"
	"net/http"
)

func HandleMaigretSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := "https://api.maigret.xyz/search?username=" + username
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	return success(string(body))
}