package tools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HandleAddIpfs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, _ :=getString(args, "data")
	if data == "" {
		return err("data is required")
}

	resp, e := http.DefaultClient.Post("http://localhost:5001/api/v0/block/put", "application/octet-stream", strings.NewReader(data))
	if e != nil {
		return err("Failed to connect to IPFS: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("Failed to parse response: " + e.Error())
}

	cid, found := result["Key"].(string)
	if !found {
		return err("Unexpected response format")
}

	return success("Added to IPFS with CID: " + cid)
}

func HandleCatIpfs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cid, _ :=getString(args, "cid")
	if cid == "" {
		return err("cid is required")
}

	u := "http://localhost:5001/api/v0/cat?arg=" + url.QueryEscape(cid)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("Failed to fetch from IPFS: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return success(string(body))
}