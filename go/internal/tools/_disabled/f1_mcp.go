package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCurrentStandings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://ergast.com/api/f1/current/driverStandings.json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch standings: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(string(body))
}