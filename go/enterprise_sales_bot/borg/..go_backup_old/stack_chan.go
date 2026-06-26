package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type stackQuestion struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func HandleGetRecentStackQuestions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tag, _ :=getString(args, "tag")
	u := fmt.Sprintf("https://api.stackexchange.com/2.3/questions?order=desc&sort=activity&tagged=%s&site=stackoverflow&pagesize=5", url.QueryEscape(tag))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch questions")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result struct {
		Items []stackQuestion `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	if len(result.Items) == 0 {
		return success("No questions found")
}

	return ok(fmt.Sprintf("Recent questions for tag '%s':\n", tag) + formatQuestions(result.Items))
}

func formatQuestions(qs []stackQuestion) string {
	var s string
	for _, q := range qs {
		s += fmt.Sprintf("- %s (%s)\n", q.Title, q.Link)

	return s
}

}

func HandleGetStackUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userId, _ :=getString(args, "userId")
	u := fmt.Sprintf("https://api.stackexchange.com/2.3/users/%s?order=desc&sort=reputation&site=stackoverflow", url.QueryEscape(userId))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch user")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result struct {
		Items []struct {
			DisplayName string `json:"display_name"`
			Reputation  int    `json:"reputation"`
			Link        string `json:"link"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	if len(result.Items) == 0 {
		return success("User not found")
}

	user := result.Items[0]
	msg := fmt.Sprintf("User: %s\nReputation: %d\nProfile: %s", user.DisplayName, user.Reputation, user.Link)
	return ok(msg)
}