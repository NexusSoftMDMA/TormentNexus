package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleWolframAlpha(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	appID := os.Getenv("WOLFRAM_ALPHA_APPID")
	if appID == "" {
		return err("WOLFRAM_ALPHA_APPID not set")
}

	u := fmt.Sprintf("https://api.wolframalpha.com/v2/query?input=%s&appid=%s&output=json&format=plaintext", url.QueryEscape(query), appID)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	var result struct {
		QueryResult struct {
			Pods []struct {
				Title    string `json:"title"`
				Subpods []struct {
					Plaintext string `json:"plaintext"`
				} `json:"subpods"`
			} `json:"pods"`
		} `json:"queryresult"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	if len(result.QueryResult.Pods) == 0 || len(result.QueryResult.Pods[0].Subpods) == 0 {
		return err("no result found")
}

	answer := result.QueryResult.Pods[0].Subpods[0].Plaintext
	if answer == "" {
		return err("empty result")
}

	return ok(answer)
}