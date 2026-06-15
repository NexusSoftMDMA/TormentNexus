package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleTeam(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	team, _ :=getString(args, "team")
	if team == "" {
		return err("missing 'team' argument")
}

	url := fmt.Sprintf("https://api.statbotics.io/v2/team/%s", team)
	e := getJSON(url)
	if e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Team %s data retrieved", team))
}

func HandleEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "event_key")
	if key == "" {
		return err("missing 'event_key' argument")
}

	url := fmt.Sprintf("https://api.statbotics.io/v2/event/%s", key)
	e := getJSON(url)
	if e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Event %s data retrieved", key))
}

func getJSON(url string) error {
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return fmt.Errorf("HTTP request failed: %w", e)
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

	return nil
}