package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetLeetcodeProblem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	titleSlug, _ :=getString(args, "titleSlug")
	if titleSlug == "" {
		return err("titleSlug is required")
}

	query := `{"query":"query getQuestionDetail($titleSlug: String!) { question(titleSlug: $titleSlug) { title content difficulty } }","variables":{"titleSlug":"` + titleSlug + `"}}`
	resp, e := http.DefaultClient.Post("https://leetcode.com/graphql", "application/json", bytes.NewReader([]byte(query)))
	if e != nil {
		return err("failed to fetch problem: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("unexpected response format")
}

	question, found := data["question"].(map[string]interface{})
	if !found {
		return err("problem not found")
}

	jsonBytes, e := json.Marshal(question)
	if e != nil {
		return err("failed to marshal result")
}

	return ok(string(jsonBytes))
}