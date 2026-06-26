package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	reqURL := "http://localhost:3000/users/" + id
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("user not found")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}

func HandleCreateUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	email, _ :=getString(args, "email")
	if name == "" || email == "" {
		return err("name and email are required")
}

	payload := map[string]string{"name": name, "email": email}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	resp, e := http.Post("http://localhost:3000/users", "application/json", strings.NewReader(string(data)))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return err("creation failed")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}