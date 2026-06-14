package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListEmployees(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/hr/employees")
	if e != nil {
		return err("failed to fetch employees: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetEmployee(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://api.example.com/hr/employees/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch employee: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}