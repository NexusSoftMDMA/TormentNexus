package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSavePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing url parameter")
}

	saveURL := fmt.Sprintf("https://web.archive.org/save/%s", url.QueryEscape(urlStr))
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, saveURL, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return err(fmt.Sprintf("save failed, status: %d", resp.StatusCode))
}

	return ok("page successfully submitted for archiving")
}

func HandleCheckArchive(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing url parameter")
}

	apiURL := fmt.Sprintf("https://archive.org/wayback/available?url=%s", url.QueryEscape(urlStr))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result struct {
		ArchivedSnapshots map[string]struct {
			Status string `json:"status"`
		} `json:"archived_snapshots"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json decode: %v", e))
}

	if snap, found := result.ArchivedSnapshots["closest"]; found && snap.Status == "200" {
		return ok("page is archived")
}

	return ok("no archive found")
}// touch 1781132128
