package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	token, _ :=getString(args, "token")
	instance, _ :=getString(args, "instanceUrl")

	req, e := http.NewRequestWithContext(ctx, "GET", instance+"/services/data/v58.0/query/?q="+query, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("query failed")
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}

func HandleDescribe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	object, _ :=getString(args, "object")
	token, _ :=getString(args, "token")
	instance, _ :=getString(args, "instanceUrl")

	req, e := http.NewRequestWithContext(ctx, "GET", instance+"/services/data/v58.0/sobjects/"+object+"/describe", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("describe failed")
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}