package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSubxtGetMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	payload := fmt.Sprintf(`{"id":1,"jsonrpc":"2.0","method":"state_getMetadata","params":[]}`)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if errMsg, found := result["error"]; found {
		return err(fmt.Sprintf("rpc error: %v", errMsg))
}

	metadata, found := result["result"]
	if !found {
		return err("no result in response")
}

	return success(fmt.Sprintf("Metadata: %v", metadata))
}

func HandleSubxtGetBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	blockNumber, _ :=getString(args, "blockNumber")
	var params string
	if blockNumber == "" {
		params = "[]"
	} else {
		params = fmt.Sprintf(`["%s"]`, blockNumber)

	payload := fmt.Sprintf(`{"id":1,"jsonrpc":"2.0","method":"chain_getBlock","params":%s}`, params)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if errMsg, found := result["error"]; found {
		return err(fmt.Sprintf("rpc error: %v", errMsg))
}

	block, found := result["result"]
	if !found {
		return err("no result in response")
}

	return success(fmt.Sprintf("Block: %v", block))
}
}