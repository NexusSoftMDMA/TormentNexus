package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleGetCurrentTime_time_node_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	timezone, _ :=getString(args, "timezone")
	loc := time.UTC
	if timezone != "" {
		var e error
		loc, e = time.LoadLocation(timezone)
		if e != nil {
			return err("invalid timezone: " + e.Error())

	}
	now := time.Now().In(loc)
	data, _ := json.Marshal(map[string]string{"time": now.Format(time.RFC3339), "timezone": loc.String()})
	return success(string(data))
}

}

func HandleConvertTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fromTz, _ :=getString(args, "from_timezone")
	toTz, _ :=getString(args, "to_timezone")
	timeStr, _ :=getString(args, "time")
	locFrom, e := time.LoadLocation(fromTz)
	if e != nil {
		return err("invalid from_timezone: " + e.Error())
}

	locTo, e := time.LoadLocation(toTz)
	if e != nil {
		return err("invalid to_timezone: " + e.Error())
}

	t, e := time.ParseInLocation(time.RFC3339, timeStr, locFrom)
	if e != nil {
		return err("invalid time format: " + e.Error())
}

	converted := t.In(locTo)
	data, _ := json.Marshal(map[string]string{"converted_time": converted.Format(time.RFC3339), "timezone": toTz})
	return success(string(data))
}

func init_time_node_mcp() {
	http.DefaultClient.Timeout = 10 * time.Second
}