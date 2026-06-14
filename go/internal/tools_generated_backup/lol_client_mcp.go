package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetSummoner(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("summoner name is required")
}

	url := fmt.Sprintf("http://localhost:2999/lol-summoner/v1/summoners?name=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	return ok(data)
}

func HandleGetChampSelect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "http://localhost:2999/lol-champ-select/v1/session"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	return ok(data)
}