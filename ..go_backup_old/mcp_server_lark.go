package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://open.larksuite.com/open-apis/search/v1?query=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Lark search API: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	msg, found := result["message"]
	if !found {
		return success("search completed, no message field")
}

	return success(fmt.Sprintf("%v", msg))
}