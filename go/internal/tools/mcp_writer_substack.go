//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleWriteSubstackPost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	body, _ :=getString(args, "body")
	token, _ :=getString(args, "token")
	pub, _ :=getString(args, "publication")

	if title == "" || body == "" || token == "" || pub == "" {
		return err("Missing required args: title, body, token, publication")
}

	payload := fmt.Sprintf(`{"title":"%s","body":"%s","publication":"%s"}`, title, strings.ReplaceAll(body, `"`, `\"`), pub)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.substack.com/api/v1/posts", strings.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("Failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Request failed: %v", e))
}

	defer resp.Body.Close()

	bodyBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read response: %v", e))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes)))
	return ok("Post written successfully")
}

func HandleListSubstackPublications(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("Missing required arg: token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.substack.com/api/v1/publications", nil)
	if e != nil {
		return err(fmt.Sprintf("Failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Request failed: %v", e))
}

	defer resp.Body.Close()

	bodyBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read response: %v", e))
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(bodyBytes)))
}

	var data []map[string]interface{	if e := json.Unmarshal(bodyBytes, &data); e != nil {
		return err(fmt.Sprintf("Failed to parse JSON: %v", e))
}

	var names []string
	for _, pub := range data {
		if name, found := pub["name"].(string); found {
			names = append(names, name)

		return ok(strings.Join(names, ", "))
}
}
}
}
}
}
}