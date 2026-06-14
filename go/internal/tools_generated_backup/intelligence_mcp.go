package tools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/intelligence?q=" + query)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleProcess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	threshold, _ :=getInt(args, "threshold")
	if threshold == 0 {
		threshold = 50
	}
	if len(text) > threshold {
		return success("text length exceeds threshold")
}

	return ok("processed: " + text)
}