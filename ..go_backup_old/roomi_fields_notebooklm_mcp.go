package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	notebookID, _ :=getString(args, "notebook_id")
	query, _ :=getString(args, "query")
	if notebookID == "" || query == "" {
		return err("notebook_id and query are required")
}

	url := fmt.Sprintf("https://api.notebooklm.google.com/v1/notebooks/%s/query?query=%s", notebookID, query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Result: %v", result))
}