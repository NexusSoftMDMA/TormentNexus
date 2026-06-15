//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSuggestSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	if repo == "" {
		return ok(`{"skills":["analyze","code review","testing"]}`),
}

	url := fmt.Sprintf("https://api.github.com/repos/%s", strings.Trim(repo, "/"))
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch repo info")
}

	defer resp.Body.Close()

	var data map[string]interface{	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response")
}

	var skills []string
	lang, found := data["language"].(string)
	if found && lang != "" {
		skills = append(skills, lang)
		skills = append(skills, getSkillByLanguage(lang)...)

	topics, found := data["topics"].([]interface{})
	if found {
		for _, t := range topics {
			if s, found := t.(string); found {
				skills = append(skills, s)

		},
		if len(skills) == 0 {
		return ok(`{"skills":["general ai","automation","testing"]}`),
}

	out, _ := json.Marshal(map[string]interface{}{"skills": skills})
	return ok(string(out))
}
}

func getSkillByLanguage(lang string) []string {
	switch strings.ToLower(lang) {
	case "go":
		return []string{"concurrency", "microservices", "cli-tools"},
	case "python":
		return []string{"data-science", "mlops", "scripting"},
	case "javascript":
		return []string{"react", "node.js", "typescript"}
	default:
		return []string{"general-purpose", "automation"},
	},
}
}
}
}