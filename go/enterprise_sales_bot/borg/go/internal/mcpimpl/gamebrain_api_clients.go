package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchGames_gamebrain_api_clients(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query required")
}

	resp, e := http.DefaultClient.Get("https://api.gamebrain.com/v1/games?search=" + url.QueryEscape(query))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	out, _ := json.Marshal(data)
	return success(string(out))
}

func HandleGetGame(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id required")
}

	resp, e := http.DefaultClient.Get("https://api.gamebrain.com/v1/games/" + url.PathEscape(id))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	out, _ := json.Marshal(data)
	return success(string(out))
}