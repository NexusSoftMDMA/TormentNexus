package mcpimpl

import (
	"context"
	"io/ioutil"
	"net/http"
)

func HandleGetDocumentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url argument")
}

	res, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return err("unexpected status code")
}

	body, e := ioutil.ReadAll(res.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}