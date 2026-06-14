package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type contributionsResponse struct {
	Total          int `json:"totalContributions"`
	TotalLanguages int `json:"totalLanguages"`
}

func HandleGetContributions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		username = "erickwendel"
	}
	year, _ :=getInt(args, "year")
	if year == 0 {
		year = 2024
	}
	url := fmt.Sprintf("https://github-contributions.vercel.app/api/v1/%s?year=%d", username, year)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch contributions")
}

	defer resp.Body.Close()
	var data contributionsResponse
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("Erick Wendel had %d contributions in %d", data.Total, year))
}