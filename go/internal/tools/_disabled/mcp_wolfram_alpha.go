package tools

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	appid, _ :=getString(args, "appid")
	if appid == "" {
		return err("appid is required")
}

	queryURL := fmt.Sprintf("https://api.wolframalpha.com/v2/query?input=%s&appid=%s&format=plaintext",
		url.QueryEscape(input), appid)
	req, e := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	type QueryResult struct {
		Pods []struct {
			Title    string `xml:"title,attr"`
			SubPods []struct {
				Plaintext string `xml:"plaintext"`
			} `xml:"subpod"`
		} `xml:"pod"`
	}
	var result QueryResult
	if e := xml.Unmarshal(body, &result); e != nil {
		return err("failed to parse XML: " + e.Error())
}

	output := ""
	for _, pod := range result.Pods {
		for _, sub := range pod.SubPods {
			if sub.Plaintext != "" {
				output += fmt.Sprintf("%s: %s\n", pod.Title, sub.Plaintext)

		}
	}
	if output == "" {
		output = "No plaintext results found."
	}
	return ok(output)
}
}