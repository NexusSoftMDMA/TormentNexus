package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetFlight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flightIATA, _ :=getString(args, "flight_iata")
	if flightIATA == "" {
		return err("flight_iata is required")
}

	accessKey := os.Getenv("AVIATIONSTACK_ACCESS_KEY")
	if accessKey == "" {
		return err("AVIATIONSTACK_ACCESS_KEY not set")
}

	url := fmt.Sprintf("http://api.aviationstack.com/v1/flights?access_key=%s&flight_iata=%s", accessKey, flightIATA)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(result)
}