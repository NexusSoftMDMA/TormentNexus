package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	anthropicEndpoint = "https://api.anthropic.com/v1/messages"
	anthropicVersion  = "2023-06-01"
	model             = "claude-haiku-4-5-20251001"
)

type TextContent struct {
	Text string `json:"text"`
}

func HandleClassify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemNumStr, _ :=getString(args, "item_num")
	itemNum, _ :=getInt(args, "item_num")
	if e != nil {
		return err(e.Error())
}

	item := map[string]interface{}{
		"title": getString(args, "title"),
		"body":  getString(args, "body"),
		"labels": args["labels"],
	}

	classifiedLabels, classifyErr := classify(item, itemNum)
	if classifyErr != nil {
		return err(classifyErr.Error())
}

	return ok(classifiedLabels)
}

func classify(item map[string]interface{}, itemNum int) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY env var is required")
}

	existing := []string{}
	if labels, found := item["labels"].([]interface{}); found {
		for _, lbl := range labels {
			if label, found := lbl.(map[string]interface{}); found {
				if name, found := label["name"].(string); found {
					existing = append(existing, name)

			}
		}
	}

	body, found := item["body"].(string)
	if !found {
		body = "(empty)"
	} else {
		body = strings.TrimSpace(body)
		if body == "" {
			body = "(empty)"
		}
	}

	title, found := item["title"].(string)
	if !found {
		title = ""
	}

	userMsg := fmt.Sprintf("Item: #%d\nTitle: %s\nExisting labels: %s\n\nBody:\n%s", itemNum, title, strings.Join(existing, ", "), body)

	payload := map[string]interface{}{
		"model":      model,
		"max_tokens": 256,
		"system":     "You auto-label GitHub issues and PRs for the lemonade-sdk/lemonade repository.",
		"messages": []map[string]interface{}{
			{"role": "user", "content": userMsg},
		},
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return "", marshalErr
	}

	req, reqErr := http.NewRequest("POST", anthropicEndpoint, strings.NewReader(string(payloadBytes)))
	if reqErr != nil {
		return "", reqErr
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return "", fetchErr
	}
	defer resp.Body.Close()

	respBody, readErr := http.NewResponseController(resp).ReadAll(resp.Body)
	if readErr != nil {
		return "", readErr
	}

	var response map[string]interface{}
	if jsonErr := json.Unmarshal(respBody, &response); jsonErr != nil {
		return "", jsonErr
	}

	content, found := response["content"].(string)
	if !found {
		return "", fmt.Errorf("invalid response content")
}

	return content, nil
}

}

func HandleGhView(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	num, _ :=getString(args, "num")
	repo, _ :=getString(args, "repo")
	cmd := exec.Command("gh", "issue", "view", num, "--json", "title,body,labels,url")
	if repo != "" {
		cmd.Args = append(cmd.Args, "--repo", repo)

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to execute command: %s", execErr.Error()))
}

	return ok(string(output))
}
}