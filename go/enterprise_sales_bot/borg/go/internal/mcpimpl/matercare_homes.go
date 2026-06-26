package mcpimpl

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleMatercareHomes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://api.matercare.com/homes?q=" + query
	resp, e := http.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}