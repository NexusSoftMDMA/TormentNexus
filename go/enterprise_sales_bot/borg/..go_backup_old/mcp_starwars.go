package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetPerson(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	url := fmt.Sprintf("https://swapi.dev/api/people/%d/", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch person: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("person not found")
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	name, found := data["name"].(string)
	if !found {
		return err("missing name")
}

	return ok(fmt.Sprintf("Person: %s", name))
}

func HandleGetFilm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	url := fmt.Sprintf("https://swapi.dev/api/films/%d/", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch film: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("film not found")
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	title, found := data["title"].(string)
	if !found {
		return err("missing title")
}

	return ok(fmt.Sprintf("Film: %s", title))
}