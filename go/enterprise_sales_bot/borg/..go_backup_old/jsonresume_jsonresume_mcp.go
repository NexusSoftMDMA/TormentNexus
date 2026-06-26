package tools

import (
	"context"
	"encoding/json"
)

func HandleResumeGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sample := `{"$schema":"https://jsonresume.org/schema","basics":{"name":"John Doe","label":"Programmer","email":"john@example.com"}}`
	return ok(sample)
}

func HandleResumeValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resumeStr, _ :=getString(args, "resume")
	if resumeStr == "" {
		return err("resume argument is required")
}

	var found interface{}
	e := json.Unmarshal([]byte(resumeStr), &found)
	if e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("resume is valid JSON")
}