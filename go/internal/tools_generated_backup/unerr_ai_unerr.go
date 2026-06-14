package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleGetCallGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file")
	if filePath == "" {
		return err("missing file")
	}
	u := "http://localhost:9090/callgraph?file=" + url.QueryEscape(filePath)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request error: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: "+e.Error())
	}
	return success(string(body))
}

func HandleGetRules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ruleName, _ :=getString(args, "rule")
	if ruleName == "" {
		return err("missing rule")
	}
	u := "http://localhost:9090/rules?name=" + url.QueryEscape(ruleName)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request error: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: "+e.Error())
	}
	return success(string(body))
}// touch 1781132143
