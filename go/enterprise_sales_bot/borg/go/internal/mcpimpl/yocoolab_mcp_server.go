package mcpimpl

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetFeedbackThreads(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	threadID, _ :=getString(args, "thread_id")
	if threadID == "" {
		return err("thread_id is required")
}

	url := fmt.Sprintf("https://api.yocoolab.com/threads/%s", threadID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return success(string(body))
}

func GetDesignSelections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	url := fmt.Sprintf("https://api.yocoolab.com/projects/%s/selections", projectID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return success(string(body))
}