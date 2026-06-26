package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// HandleDetectMarketingTech fetches a URL and checks for common marketing technology scripts.
func HandleDetectMarketingTech(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	content := string(body)
	patterns := []string{"google-analytics", "gtag", "fbq", "gtm", "segment"}
	foundTags := []string{}
	for _, p := range patterns {
		if strings.Contains(content, p) {
			foundTags = append(foundTags, p)

	}
	return ok("Detected marketing technologies: " + strings.Join(foundTags, ", "))
}

}

// HandleExtractAnalyticsTags extracts analytics tag patterns from a given URL.
func HandleExtractAnalyticsTags(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	content := string(body)
	analyticsTags := []string{}
	patterns := []string{"GA_MEASUREMENT_ID", "UA-", "G-", "AW-", "DC-"}
	for _, p := range patterns {
		if strings.Contains(content, p) {
			analyticsTags = append(analyticsTags, p)

	}
	return ok("Analytics tags found: " + strings.Join(analyticsTags, ", "))
}
}