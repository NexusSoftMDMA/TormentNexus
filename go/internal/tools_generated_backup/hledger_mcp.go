package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleHledgerQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getString(args, "port")
	if port == "" {
		port = "5000"
	}
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query required")
}

	url := fmt.Sprintf("http://%s:%s/?query=%s", host, port, strings.ReplaceAll(query, " ", "%20"))
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}