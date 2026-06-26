package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func HandleDeepContextList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	root, _ :=getString(args, "root")
	if root == "" {
		root = "."
	}

	var contexts []string
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".dc.json") {
			contexts = append(contexts, strings.TrimSuffix(info.Name(), ".dc.json"))

		return nil
	})

	if walkErr != nil {
		return err(fmt.Sprintf("failed to walk directory: %v", walkErr))
}

	return ok(strings.Join(contexts, "\n"))
}

}

func HandleDeepContextGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("context name is required")
}

	root, _ :=getString(args, "root")
	if root == "" {
		root = "."
	}

	filePath := filepath.Join(root, name+".dc.json")
	data, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read context file: %v", readErr))
}

	var contextData map[string]interface{}
	parseErr := json.Unmarshal(data, &contextData)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse context file: %v", parseErr))
}

	jsonData, marshalErr := json.MarshalIndent(contextData, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal context data: %v", marshalErr))
}

	return ok(string(jsonData))
}

func HandleDeepContextSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("context name is required")
}

	root, _ :=getString(args, "root")
	if root == "" {
		root = "."
	}

	data, _ :=getString(args, "data")
	if data == "" {
		return err("context data is required")
}

	var contextData map[string]interface{}
	parseErr := json.Unmarshal([]byte(data), &contextData)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid JSON data: %v", parseErr))
}

	jsonData, marshalErr := json.MarshalIndent(contextData, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal context data: %v", marshalErr))
}

	filePath := filepath.Join(root, name+".dc.json")
	writeErr := os.WriteFile(filePath, jsonData, 0644)
	if writeErr != nil {
		return err(fmt.Sprintf("failed to write context file: %v", writeErr))
}

	return ok("context saved successfully")
}

func HandleDeepContextDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("context name is required")
}

	root, _ :=getString(args, "root")
	if root == "" {
		root = "."
	}

	filePath := filepath.Join(root, name+".dc.json")
	removeErr := os.Remove(filePath)
	if removeErr != nil {
		return err(fmt.Sprintf("failed to delete context file: %v", removeErr))
}

	return ok("context deleted successfully")
}

func HandleDeepContextValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, _ :=getString(args, "data")
	if data == "" {
		return err("context data is required")
}

	var contextData map[string]interface{}
	parseErr := json.Unmarshal([]byte(data), &contextData)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid JSON data: %v", parseErr))
}

	if _, found := contextData["name"]; !ok {
		return err("context must have a 'name' field")
}

	if _, found := contextData["version"]; !ok {
		return err("context must have a 'version' field")
}

	return ok("context is valid")
}

func HandleDeepContextMerge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base")
	if base == "" {
		return err("base context data is required")
}

	overlay, _ :=getString(args, "overlay")
	if overlay == "" {
		return err("overlay context data is required")
}

	var baseData, overlayData map[string]interface{}
	baseErr := json.Unmarshal([]byte(base), &baseData)
	if baseErr != nil {
		return err(fmt.Sprintf("invalid base JSON data: %v", baseErr))
}

	overlayErr := json.Unmarshal([]byte(overlay), &overlayData)
	if overlayErr != nil {
		return err(fmt.Sprintf("invalid overlay JSON data: %v", overlayErr))
}

	merged := make(map[string]interface{})
	for k, v := range baseData {
		merged[k] = v
	}
	for k, v := range overlayData {
		merged[k] = v
	}

	jsonData, marshalErr := json.MarshalIndent(merged, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal merged data: %v", marshalErr))
}

	return ok(string(jsonData))
}