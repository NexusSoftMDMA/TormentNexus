package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGetCharacter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	base := "https://api.onepieceoracle.com/characters"
	u, e := url.Parse(base)
	if e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	q := u.Query()
	q.Set("name", name)
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("Character: %v", result))
}

func HandleGetFruit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	base := "https://api.onepieceoracle.com/fruits"
	u, e := url.Parse(base)
	if e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	q := u.Query()
	q.Set("name", name)
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("Fruit: %v", result))
}