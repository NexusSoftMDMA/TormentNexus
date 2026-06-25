package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var http.DefaultClient = http.DefaultClient

func HandleListTaskLists(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	req, reqErr := http.NewRequestWithContext(ctx, "GET", "https://tasks.googleapis.com/tasks/v1/users/@me/lists", nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, resp.Status))
}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Updated string `json:"updated"`
		} `json:"items"`
	}
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var sb strings.Builder
	sb.WriteString("Task Lists:\n")
	for _, list := range result.Items {
		sb.WriteString(fmt.Sprintf("- ID: %s, Title: %s, Updated: %s\n", list.ID, list.Title, list.Updated))

	return ok(sb.String())
}

}

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	listID, _ :=getString(args, "list_id")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	if listID == "" {
		return err("missing required parameter: list_id")
}

	apiURL := fmt.Sprintf("https://tasks.googleapis.com/tasks/v1/lists/%s/tasks", listID)
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, resp.Status))
}

	var result struct {
		Items []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
			Due    string `json:"due"`
			Notes  string `json:"notes"`
		} `json:"items"`
	}
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tasks in list %s:\n", listID))
	for _, task := range result.Items {
		sb.WriteString(fmt.Sprintf("- ID: %s, Title: %s, Status: %s, Due: %s\n", task.ID, task.Title, task.Status, task.Due))
		if task.Notes != "" {
			sb.WriteString(fmt.Sprintf("  Notes: %s\n", task.Notes))

	}
	return ok(sb.String())
}

}

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	listID, _ :=getString(args, "list_id")
	title, _ :=getString(args, "title")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	if listID == "" {
		return err("missing required parameter: list_id")
}

	if title == "" {
		return err("missing required parameter: title")
}

	reqBody := map[string]string{
		"title": title,
	}
	if notes := getString(args, "notes"); notes != "" {
		reqBody["notes"] = notes
	}
	if dueDate := getString(args, "due_date"); dueDate != "" {
		reqBody["due"] = dueDate
	}

	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	apiURL := fmt.Sprintf("https://tasks.googleapis.com/tasks/v1/lists/%s/tasks", listID)
	req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(jsonBody)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, resp.Status))
}

	var result struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("Task created successfully: ID=%s, Title=%s, Status=%s", result.ID, result.Title, result.Status))
}

func HandleDeleteTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	taskID, _ :=getString(args, "task_id")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	if taskID == "" {
		return err("missing required parameter: task_id")
}

	apiURL := fmt.Sprintf("https://tasks.googleapis.com/tasks/v1/tasks/%s", taskID)
	req, reqErr := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return err(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, resp.Status))
}

	return ok(fmt.Sprintf("Task %s deleted successfully", taskID))
}