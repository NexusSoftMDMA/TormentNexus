package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetDefinition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filePath")
	line, _ :=getInt(args, "line")
	column, _ :=getInt(args, "column")
	url := fmt.Sprintf("http://localhost:1234/definition?file=%s&line=%d&col=%d", filePath, line, column)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to get definition: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("Failed to parse response")
}

	def, found := result["definition"]
	if !found {
		return err("No definition found")
}

	return ok(fmt.Sprintf("%v", def))
}

func HandleGetReferences(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filePath")
	line, _ :=getInt(args, "line")
	column, _ :=getInt(args, "column")
	url := fmt.Sprintf("http://localhost:1234/references?file=%s&line=%d&col=%d", filePath, line, column)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to get references: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("Failed to parse response")
}

	refs, found := result["references"]
	if !found {
		return err("No references found")
}

	return ok(fmt.Sprintf("%v", refs))
}