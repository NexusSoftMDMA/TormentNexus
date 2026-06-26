package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandlePowerBIQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups?query=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return success("PowerBI data: " + string(body))
}