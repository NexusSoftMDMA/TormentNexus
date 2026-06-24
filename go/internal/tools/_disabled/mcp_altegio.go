package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetAppointments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	url := "https://api.altegio.com/appointments?location=" + location
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch appointments")
}

	defer resp.Body.Close()

	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	result, _ := json.Marshal(data)
	return ok(string(result))
}