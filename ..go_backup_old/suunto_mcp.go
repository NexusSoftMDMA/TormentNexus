package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSuuntoRecent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.suunto.com/v1/activities?limit=5", nil)
	if e != nil {
		return err("request creation failed")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api call failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed")
}

	activities, found := data["activities"].([]interface{})
	if !found {
		return err("no activities found")
}

	return ok(fmt.Sprintf("Got %d activities", len(activities)))
}