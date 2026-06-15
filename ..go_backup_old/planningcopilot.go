package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListPlans(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.planningcopilot.com/plans")
	if e != nil {
		return err("failed to get plans: " + e.Error())
}

	defer resp.Body.Close()
	var plans []string
	if e := json.NewDecoder(resp.Body).Decode(&plans); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success("Plans: " + fmt.Sprintf("%v", plans))
}

func HandleCreatePlan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	body, e := json.Marshal(map[string]string{"title": title})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.planningcopilot.com/plans", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create plan: " + e.Error())
}

	defer resp.Body.Close()
	return success("Plan created: " + title)
}