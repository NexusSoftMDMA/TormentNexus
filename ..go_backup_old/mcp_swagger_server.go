package tools

import (
	"context"
	"io/ioutil"
	"net/http"
)

func HandleFetchSwagger(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch swagger: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}