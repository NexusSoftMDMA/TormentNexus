package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

var subs = make(map[string]string)

func HandleListSubscriptions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, e := json.Marshal(subs)
	if e != nil {
		return err("failed to marshal subscriptions")
}

	return ok(string(data))
}

func HandleAddSubscription(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	status, _ :=getString(args, "status")
	if name == "" {
		return err("name is required")
}

	subs[name] = status
	return success(fmt.Sprintf("added subscription %s", name))
}