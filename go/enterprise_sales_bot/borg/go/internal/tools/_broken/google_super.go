package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	googleSearchRe = regexp.MustCompile(`<div class="g">.*?<h3 class="r"><a href="(.*?)"`)
)

func HandleGoogleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, reqErr := client.Get(fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to fetch Google search results: %v", reqErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	matches := googleSearchRe.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		return err("no search results found")
}

	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])

	}

	return ok(strings.Join(results, "\n"))
}

}

func HandleGoogleTranslate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	fromLang, _ :=getString(args, "from")
	toLang, _ :=getString(args, "to")
	if fromLang == "" || toLang == "" {
		return err("from and to language parameters are required")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, reqErr := client.Get(fmt.Sprintf("https://translate.google.com/m?sl=%s&tl=%s&q=%s", fromLang, toLang, url.QueryEscape(text)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to fetch translation: %v", reqErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Simple extraction - in real implementation we'd parse the HTML properly
	translation := strings.TrimSpace(strings.SplitAfter(string(body), "class=\"result-container\">")[1])
	translation = strings.Split(translation, "</div>")[0]
	translation = strings.ReplaceAll(translation, "<div>", "")
	translation = strings.ReplaceAll(translation, "</div>", "")

	if translation == "" {
		return err("translation failed")
}

	return ok(translation)
}

func HandleGoogleWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	if location == "" {
		return err("location parameter is required")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, reqErr := client.Get(fmt.Sprintf("https://www.google.com/search?q=weather+%s", url.QueryEscape(location)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to fetch weather data: %v", reqErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Simple extraction - in real implementation we'd parse the HTML properly
	weather := strings.TrimSpace(strings.SplitAfter(string(body), "class=\"BNeawe iBp4i AP7Wnd\">")[1])
	weather = strings.Split(weather, "</div>")[0]

	if weather == "" {
		return err("weather data not found")
}

	return ok(weather)
}

func HandleGoogleCalculator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	expression, _ :=getString(args, "expression")
	if expression == "" {
		return err("expression parameter is required")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, reqErr := client.Get(fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(expression)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to fetch calculation result: %v", reqErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Simple extraction - in real implementation we'd parse the HTML properly
	result := strings.TrimSpace(strings.SplitAfter(string(body), "class=\"a61j6\">")[1])
	result = strings.Split(result, "</div>")[0]

	if result == "" {
		return err("calculation result not found")
}

	return ok(result)
}

func HandleGoogleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	if location == "" {
		return err("location parameter is required")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, reqErr := client.Get(fmt.Sprintf("https://www.google.com/search?q=time+in+%s", url.QueryEscape(location)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to fetch time data: %v", reqErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Simple extraction - in real implementation we'd parse the HTML properly
	timeStr := strings.TrimSpace(strings.SplitAfter(string(body), "class=\"BNeawe iBp4i AP7Wnd\">")[1])
	timeStr = strings.Split(timeStr, "</div>")[0]

	if timeStr == "" {
		return err("time data not found")
}

	return ok(timeStr)
}

func HandleGoogleCLI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command parameter is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("command failed: %v\nOutput: %s", runErr, string(output)))
}

	return ok(string(output))
}