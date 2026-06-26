package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleEnhanceResume(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resumeStr, _ :=getString(args, "resume_json")
	if resumeStr == "" {
		return err("resume_json is required")
}

	skill, _ :=getString(args, "skill")
	if skill == "" {
		return err("skill is required")
}

	var resume map[string]interface{}
	if e := json.Unmarshal([]byte(resumeStr), &resume); e != nil {
		return err("invalid resume_json: " + e.Error())
}

	skills, found := resume["skills"].([]interface{})
	if !found {
		skills = []interface{}{}
	}
	skills = append(skills, skill)
	resume["skills"] = skills
	out, e := json.Marshal(resume)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	return ok(string(out))
}

func HandleValidateResume(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resumeStr, _ :=getString(args, "resume_json")
	if resumeStr == "" {
		return err("resume_json is required")
}

	var resume map[string]interface{}
	if e := json.Unmarshal([]byte(resumeStr), &resume); e != nil {
		return err("invalid JSON: " + e.Error())
}

	_, found := resume["name"]
	if !found {
		return err("missing 'name' field")
}

	_, found = resume["skills"]
	if !found {
		return err("missing 'skills' field")
}

	return ok("resume is valid")
}