//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	listID, _ :=getString(args, "list_id")
	if token == "" || listID == "" {
		return err("missing token or list_id")
}

	url := fmt.Sprintf("https://api.clickup.com/api/v2/list/%s/task", listID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	


-reasoner (deepseek)*
}
}