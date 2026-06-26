package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetDailyChallenge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query := `{"query":"query { dailyChallenge { date link question { title titleSlug difficulty } } }"}`
	req, e := http.NewRequestWithContext(ctx, "POST", "https://leetcode.com/graphql", strings.NewReader(query))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("unexpected response structure")
}

	dc, found := data["dailyChallenge"].(map[string]interface{})
	if !found {
		return err("daily challenge not found")
}

	return ok(fmt.Sprintf("Daily challenge: %v", dc))
}

func HandleGetProblem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	slug, _ :=getString(args, "slug")
	if slug == "" {
		return err("slug is required")
}

	query := fmt.Sprintf(`{"query":"query questionData($titleSlug: String!) { question(titleSlug: $titleSlug) { title titleSlug difficulty content } }","variables":{"titleSlug":"%s"}}`, slug)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://leetcode.com/graphql", strings.NewReader(query))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("unexpected response structure")
}

	q, found := data["question"].(map[string]interface{})
	if !found {
		return err("problem not found")
}

	return ok(fmt.Sprintf("Problem: %v", q))
}