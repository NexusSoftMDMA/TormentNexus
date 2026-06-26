package mcpimpl

import (
	"context"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// HandleGetPageTitle fetches the <title> of a webpage.
func HandleGetPageTitle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body: " + e.Error())
}

	re := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	match := re.FindStringSubmatch(string(body))
	if len(match) < 2 {
		return err("no title found")
}

	title := strings.TrimSpace(match[1])
	return success(title)
}

// HandleGetMetaDescription fetches the meta description of a webpage.
func HandleGetMetaDescription(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body: " + e.Error())
}

	re := regexp.MustCompile(`(?i)<meta\s+name=["']description["']\s+content=["'](.*?)["']\s*/?>`)
	match := re.FindStringSubmatch(string(body))
	if len(match) < 2 {
		return err("no meta description found")
}

	desc := strings.TrimSpace(match[1])
	return success(desc)
}