package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListRegisters(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://register.openregister.gov.uk/registers")
	if e != nil {
		return err("failed to fetch registers: " + e.Error())
}

	defer resp.Body.Close()
	var result []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return success(string(data))
}

func HandleGetRegister(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("register name is required")
}

	url := fmt.Sprintf("https://register.openregister.gov.uk/registers/%s", strings.TrimSpace(name))
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch register: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return err("register not found")
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return success(string(data))
}