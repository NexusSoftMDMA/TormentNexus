package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetFlightData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flightID, _ :=getString(args, "flight_id")
	if flightID == "" {
		return err("flight_id is required")
}

	url := "https://api.airblackbox.com/flights/" + flightID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return ok(string(body))
}