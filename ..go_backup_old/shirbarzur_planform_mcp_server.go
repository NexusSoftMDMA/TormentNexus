package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "diagram_id")
	if id == "" {
		return err("diagram_id is required")
}

	url := "https://api.planform.io/diagrams/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok("fetched diagram")
}

func HandleListDiagrams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.planform.io/diagrams"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var list []map[string]interface{}
	if e := json.Unmarshal(body, &list); e != nil {
		return err("parse failed: " + e.Error())
}

	return success("listed diagrams")
}