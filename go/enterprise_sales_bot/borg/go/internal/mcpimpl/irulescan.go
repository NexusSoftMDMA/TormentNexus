package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleScan_irulescan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var data map[string]interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err(fmt.Sprintf("parse JSON failed: %v", e))
}

	return ok(fmt.Sprintf("Scan completed: %+v", data))
}

func HandleCheck_irulescan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	url := fmt.Sprintf("https://irulescan.com/api/check?domain=%s", domain)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var data map[string]interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err(fmt.Sprintf("parse JSON failed: %v", e))
}

	return ok(fmt.Sprintf("Check completed: %+v", data))
}// touch 1781132128
