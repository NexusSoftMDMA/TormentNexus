package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetRancherBinary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner := "futuretea"
	repo := "rancher-mcp-server"
	tag, _ :=getString(args, "version")
	if tag == "" {
		tag = "latest"
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/%s", owner, repo, tag)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("GitHub API returned status " + resp.Status)
}

	var release struct {
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if e := json.Unmarshal(body, &release); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "windows-arm64") {
			return ok(asset.URL)

	}
	return err("no windows/arm64 binary found in release")
}
}