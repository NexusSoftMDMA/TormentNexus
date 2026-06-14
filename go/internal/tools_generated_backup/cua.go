package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("https://api.example.com/cua/users/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch user: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return success(fmt.Sprintf("User data: %v", result))
}

func HandleCreateUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	email, _ :=getString(args, "email")
	payload := map[string]string{"name": name, "email": email}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.example.com/cua/users", "application/json", body)
	if e != nil {
		return err("failed to create user: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return err("create failed: " + string(respBody))
}

	return ok("user created successfully")
}