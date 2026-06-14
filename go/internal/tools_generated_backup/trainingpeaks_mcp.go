package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetAthlete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("access_token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.trainingpeaks.com/v1/athlete", nil)
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

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Athlete: %v", data))
}

func HandleGetWorkouts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("access_token is required")
}

	athleteID, _ :=getInt(args, "athlete_id")
	url := fmt.Sprintf("https://api.trainingpeaks.com/v1/athletes/%d/workouts", athleteID)
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

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var data []map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Workouts: %v", data))
}