package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func HandleLetsrunitRunTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
	}
	apiURL := os.Getenv("LETS_RUN_IT_API")
	if apiURL == "" {
		apiURL = "https://api.letsrunit.com/run"
	}
	reqBody, e := json.Marshal(map[string]string{"url": url})
	if e != nil {
		return err("failed to marshal request")
	}
	resp, e := http.DefaultClient.Post(apiURL, "application/json", bytes.NewReader(reqBody))
	if e != nil {
		return err("failed to send request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
	}
	respBody, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response")
	}
	return success(fmt.Sprintf("Test run result: %v", result))
}