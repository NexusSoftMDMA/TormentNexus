package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchCharacters(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
	}

	url := fmt.Sprintf("https://swapi.dev/api/people/?search=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
	}

	return ok(fmt.Sprintf("Found characters: %v", result["results"]))
}

func HandleSearchPlanets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
	}

	url := fmt.Sprintf("https://swapi.dev/api/planets/?search=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
	}

	return ok(fmt.Sprintf("Found planets: %v", result["results"]))
}// touch 1781132140
