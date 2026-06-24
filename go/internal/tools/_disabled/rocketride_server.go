package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListRockets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.rocketride.com/rockets")
	if e != nil {
		return err("failed to fetch rockets: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var rockets []string
	if e := json.Unmarshal(body, &rockets); e != nil {
		return err("failed to parse rockets: " + e.Error())
}

	return ok(fmt.Sprintf("Rockets: %v", rockets))
}

func HandleGetRocket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("https://api.rocketride.com/rockets/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch rocket: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var rocket map[string]interface{}
	if e := json.Unmarshal(body, &rocket); e != nil {
		return err("failed to parse rocket: " + e.Error())
}

	return ok(fmt.Sprintf("Rocket: %v", rocket))
}