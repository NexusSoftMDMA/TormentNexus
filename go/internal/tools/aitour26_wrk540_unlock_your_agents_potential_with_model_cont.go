//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("missing 'message' argument")
}

	return ok("You said: " + msg)
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	zone, _ :=getString(args, "timezone")
	if zone == "" {
		zone = "UTC"
	}
	loc, e := time.LoadLocation(zone)
	if e != nil {
		return err("invalid timezone: " + e.Error())
}

	now := time.Now().In(loc)
	resp := map[string]string{"time": now.Format(time.RFC1123), "zone": zone}
	data, e := json.Marshal(resp)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return success(string(data))
}
