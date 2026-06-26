package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleFetchStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://httpbin.org/get")
	if e != nil {
		return err("failed to fetch status: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return success(string(body))
}