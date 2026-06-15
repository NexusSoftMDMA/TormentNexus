package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("missing access_token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.fitbit.com/1/user/-/profile.json", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(string(body))
}

func HandleGetHeartRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("missing access_token")
}

	date, _ :=getString(args, "date")
	if date == "" {
		date = "today"
	}
	url := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/heart/date/%s/1d/1min/time/00:00/23:59.json", date)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}