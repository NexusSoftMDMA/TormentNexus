package mcpimpl

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleListSports(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.davincibets.com/sports"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch sports")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok("Sports list: " + string(body))
}

func HandleGetPrediction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sport, _ :=getString(args, "sport")
	gameID, _ :=getString(args, "gameId")
	url := "https://api.davincibets.com/predictions?gameId=" + gameID + "&sport=" + sport
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch prediction")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON")
}

	return success("Prediction: " + string(body))
}