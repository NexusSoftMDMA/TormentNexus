package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateSurvey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	desc, _ :=getString(args, "description")
	if title == "" {
		return err("title is required")
	}
	body := map[string]string{"title": title, "description": desc}
	b, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("json marshal: %v", e))
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.wjx.cn/open/v1/survey/create", bytes.NewReader(b))
	if e != nil {
		return err(fmt.Sprintf("new request: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")
	apiKey, _ :=getString(args, "apiKey")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http do: %v", e))
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(fmt.Sprintf("json decode: %v", e))
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("api error: %v", result))
	}
	return success(fmt.Sprintf("survey created: %v", result))
}

}

func HandleListSurveys(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	page, _ :=getInt(args, "page")
	if page < 1 {
		page = 1
	}
	url := fmt.Sprintf("https://api.wjx.cn/open/v1/survey/list?page=%d", page)
	if size := getInt(args, "pageSize"); size > 0 {
		url += fmt.Sprintf("&size=%d", size)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("new request: %v", e))
	}
	apiKey, _ :=getString(args, "apiKey")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http do: %v", e))
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(fmt.Sprintf("json decode: %v", e))
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("api error: %v", result))
	}
	return ok(fmt.Sprintf("surveys: %v", result))
}
}
}