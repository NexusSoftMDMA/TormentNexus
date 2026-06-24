package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleConvertToOpenAI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolDef, _ :=getString(args, "tool_definition")
	if toolDef == "" {
		return err("missing tool_definition")
}

	var def map[string]interface{}
	if e := json.Unmarshal([]byte(toolDef), &def); e != nil {
		return err("invalid tool_definition: " + e.Error())
}

	name, _ := def["name"].(string)
	desc, _ := def["description"].(string)
	params, found := def["parameters"].(map[string]interface{})
	if !found {
		params = map[string]interface{}{}
	}
	openaiTool := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        name,
			"description": desc,
			"parameters":  params,
		},
	}
	out, _ := json.Marshal(openaiTool)
	return ok(string(out))
}

func HandleConvertToClaude(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolDef, _ :=getString(args, "tool_definition")
	if toolDef == "" {
		return err("missing tool_definition")
}

	var def map[string]interface{}
	if e := json.Unmarshal([]byte(toolDef), &def); e != nil {
		return err("invalid tool_definition: " + e.Error())
}

	name, _ := def["name"].(string)
	desc, _ := def["description"].(string)
	params, found := def["parameters"].(map[string]interface{})
	if !found {
		params = map[string]interface{}{}
	}
	claudeTool := map[string]interface{}{
		"name":        name,
		"description": desc,
		"input_schema": params,
	}
	out, _ := json.Marshal(claudeTool)
	return ok(string(out))
}