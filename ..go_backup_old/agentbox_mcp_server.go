package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleCreateSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	headless, _ :=getBool(args, "headless")
	payload := map[string]bool{"headless": headless}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("http://localhost:8080/api/sessions", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create session: " + e.Error())
}

	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("invalid response")
}

	sessionID, found := result["session_id"].(string)
	if !found {
		return err("no session_id in response")
}

	return ok("Session created: " + sessionID)
}

func HandleNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionID, _ :=getString(args, "session_id")
	url, _ :=getString(args, "url")
	if sessionID == "" || url == "" {
		return err("session_id and url are required")
}

	payload := map[string]string{"session_id": sessionID, "url": url}
	body, _ := json.Marshal(payload)
	reqURL := "http://localhost:8080/api/sessions/" + sessionID + "/navigate"
	resp, e := http.DefaultClient.Post(reqURL, "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to navigate: " + e.Error())
}

	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("invalid response")
}

	if status, found := result["status"].(string); found && status == "ok" {
		return ok("Navigated to " + url)
}

	return err("navigation failed")
}