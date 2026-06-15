//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success("Fetched: " + string(body))
}

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name argument is required")
}

	age, _ :=getInt(args, "age")
	msg := fmt.Sprintf("Hello, %s", name)
	if age > 0 {
		msg += fmt.Sprintf(", you are %d years old", age)

	return success(msg)
}
}
