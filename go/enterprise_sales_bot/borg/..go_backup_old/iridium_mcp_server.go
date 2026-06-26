package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleWorkout(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	url := fmt.Sprintf("https://api.iridium.fitness/workouts?user=%s", userID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch workout: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleNutrition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	url := fmt.Sprintf("https://api.iridium.fitness/nutrition?date=%s", date)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch nutrition: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}