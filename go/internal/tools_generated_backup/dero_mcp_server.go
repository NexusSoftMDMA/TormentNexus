package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCallRPC(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	method, _ :=getString(args, "method")
	params := args["params"]
	u, _ :=getString(args, "daemon_url")
	if u == "" {
		u = "http://127.0.0.1:10102/json_rpc"
	}
	reqBody, e := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(reqBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	jsonResult, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result: " + e.Error())
}

	return success(string(jsonResult))
}

func HandleListMethods(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	methods := []string{"getblockcount", "getblockhash", "getblocktemplate", "submitblock", "getlastblockheader",
		"getblockheaderbyhash", "getblockheaderbyheight", "getpeers", "getinfo", "getbalance",
		"gettransferbytxid", "gettransfers", "transfer", "transfer_split", "scinvoke",
		"scinvoke_split", "getsc", "getscstatus", "daemon_info"}
	methodsJSON, e := json.Marshal(methods)
	if e != nil {
		return err("failed to marshal methods list")
}

	return success(string(methodsJSON))
}