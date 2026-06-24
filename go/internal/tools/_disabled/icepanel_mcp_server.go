package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListDiagrams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.icepanel.io/v1/projects/"+projectID+"/diagrams", nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success("Diagrams listed successfully")
}

func HandleGetElement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	elementID, _ :=getString(args, "element_id")
	if elementID == "" {
		return err("element_id is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.icepanel.io/v1/elements/"+elementID, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success("Element retrieved successfully")
}