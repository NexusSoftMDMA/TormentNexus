package tools

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
)

func HandleLeetcodeRandomProblem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://leetcode.com/api/problems/all/")
	if e != nil {
		return err("failed to fetch problems: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data struct {
		StatStatusPairs []struct {
			Stat struct {
				FrontendQuestionID int    `json:"frontend_question_id"`
				QuestionTitle      string `json:"question__title"`
				QuestionTitleSlug  string `json:"question__title_slug"`
			} `json:"stat"`
			Difficulty struct {
				Level int `json:"level"`
			} `json:"difficulty"`
		} `json:"stat_status_pairs"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse: " + e.Error())
}

	if len(data.StatStatusPairs) == 0 {
		return err("no problems found")
}

	p := data.StatStatusPairs[rand.Intn(len(data.StatStatusPairs))]
	link := "https://leetcode.com/problems/" + p.Stat.QuestionTitleSlug
	return ok("Problem: " + p.Stat.QuestionTitle + " (ID " + string(rune(p.Stat.FrontendQuestionID)) + ") " + link)
}

func HandleLeetcodeDailyChallenge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://leetcode.com/api/problems/daily/")
	if e != nil {
		return err("failed to fetch daily: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse: " + e.Error())
}

	title, found := data["questionTitle"].(string)
	if !found {
		return err("daily challenge not available")
}

	slug, found := data["questionTitleSlug"].(string)
	if !found {
		return err("daily challenge slug missing")
}

	link := "https://leetcode.com/problems/" + slug
	return ok("Daily Challenge: " + title + " " + link)
}