package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleSearchCalls(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey required")
}

	query, _ :=getString(args, "query")
	url := "https://api.gong.io/v2/calls?query=" + query
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("api call failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(result)
}

func HandleGetTranscript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey required")
}

	callID, _ :=getString(args, "callId")
	if callID == "" {
		return err("callId required")
}

	url := "https://api.gong.io/v2/calls/" + callID + "/transcript"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("api call failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(result)
}