package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleNearbyPeople(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "lat")
	lng, _ :=getString(args, "lng")
	radius, _ :=getInt(args, "radius")
	if radius == 0 {
		radius = 1000
	}
	url := fmt.Sprintf("https://api.antenna.fyi/v1/nearby?lat=%s&lng=%s&radius=%d", lat, lng, radius)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query nearby people: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		People []string `json:"people"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d people nearby: %v", len(result.People), result.People))
}

func HandleUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.antenna.fyi/v1/profile/%s", username)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query profile: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var profile map[string]interface{}
	if e := json.Unmarshal(body, &profile); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Profile: %+v", profile))
}