package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SonarQube API client configuration
type sonarqubeClient struct {
	baseURL   string
	token     string
	org       string
	idePort   string
	projectKey string
}

func newSonarQubeClient() *sonarqubeClient {
	return &sonarqubeClient{
}
		baseURL:   getEnvOrDefault("SONARQUBE_URL", "https://sonarqube.com"),
		token:     os.Getenv("SONARQUBE_TOKEN"),
		org:       os.Getenv("SONARQUBE_ORG"),
		idePort:   os.Getenv("SONARQUBE_IDE_PORT"),
		projectKey: os.Getenv("SONARQUBE_PROJECT_KEY"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func (c *sonarqubeClient) isCloud() bool {
	return strings.Contains(c.baseURL, "sonarcloud") || c.org != ""
}

func (c *sonarqubeClient) apiURL(path string, params url.Values) string {
	base, _ := url.Parse(c.baseURL)
	base.Path = path
	if c.isCloud() && c.org != "" {
		params.Set("organization", c.org)

	base.RawQuery = params.Encode()
	return base.String()
}

func (c *sonarqubeClient) doRequest(ctx context.Context, method, path string, params url.Values, body interface{}) ([]byte, error) {
	reqURL := c.apiURL(path, params)
	
	var reqBody io.Reader
	if body != nil {
		jsonData, jsonErr := json.Marshal(body)
		if jsonErr != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", jsonErr)
}

		reqBody = bytes.NewReader(jsonData)

	req, reqErr := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)

	client := http.DefaultClient
	resp, respErr := client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("request failed: %w", respErr)
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response: %w", readErr)
}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("SonarQube API error (status %d): %s", resp.StatusCode, string(respBody))
}

	return respBody, nil
}

}
}

// HandleSearchMySonarqubeProjects searches for SonarQube projects accessible to the user
func HandleSearchMySonarqubeProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := newSonarQubeClient()
	
	if client.token == "" {
		return err("SONARQUBE_TOKEN environment variable is required")
}

	query, _ :=getString(args, "query")
	pageSize, _ :=getInt(args, "pageSize")
	if pageSize == 0 {
		pageSize = 100
	}
	page, _ :=getInt(args, "page")
	if page == 0 {
		page = 1
	}

	params := url.Values{}
	params.Set("ps", strconv.Itoa(pageSize))
	params.Set("p", strconv.Itoa(page))
	
	if query != "" {
		params.Set("q", query)

	response, apiErr := client.doRequest(ctx, "GET", "/api/projects/search", params, nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result struct {
		Paging struct {
			PageIndex int `json:"pageIndex"`
			PageSize  int `json:"pageSize"`
			Total     int `json:"total"`
		} `json:"paging"`
		Components []struct {
			Key          string `json:"key"`
			Name         string `json:"name"`
			Description  string `json:"description"`
			Visibility   string `json:"visibility"`
			LastAnalysis string `json:"lastAnalysis"`
		} `json:"components"`
	}

	parseErr := json.Unmarshal(response, &result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d projects (page %d of %d):\n\n", 
		len(result.Components), result.Paging.PageIndex, (result.Paging.Total+result.Paging.PageSize-1)/result.Paging.PageSize))
	
	for _, proj := range result.Components {
		output.WriteString(fmt.Sprintf("## %s\n", proj.Name))
		output.WriteString(fmt.Sprintf("- Key: `%s`\n", proj.Key))
		output.WriteString(fmt.Sprintf("- Visibility: %s\n", proj.Visibility))
		if proj.Description != "" {
			output.WriteString(fmt.Sprintf("- Description: %s\n", proj.Description))

		if proj.LastAnalysis != "" {
			output.WriteString(fmt.Sprintf("- Last Analysis: %s\n", proj.LastAnalysis))

		output.WriteString("\n")

	return ok(output.String())
}

}
}
}
}

// HandleGetProjectQualityGateStatus retrieves the quality gate status for a project
func HandleGetProjectQualityGateStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := newSonarQubeClient()
	
	if client.token == "" {
		return err("SONARQUBE_TOKEN environment variable is required")
}

	projectKey, _ :=getString(args, "projectKey")
	if projectKey == "" {
		return err("projectKey is required")
}

	params := url.Values{}
	params.Set("project", projectKey)

	response, apiErr := client.doRequest(ctx, "GET", "/api/qualitygates/project_status", params, nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result struct {
		ProjectStatus struct {
			Status string `json:"status"`
			Conditions []struct {
				Status       string `json:"status"`
				MetricKey    string `json:"metricKey"`
				Comparator   string `json:"comparator"`
				PeriodIndex  int    `json:"periodIndex"`
				MeasureValue string `json:"value"`
				WarningThreshold string `json:"warningThreshold"`
				ErrorThreshold string  `json:"errorThreshold"`
			} `json:"conditions"`
		} `json:"projectStatus"`
	}

	parseErr := json.Unmarshal(response, &result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("## Quality Gate Status for `%s`\n\n", projectKey))
	output.WriteString(fmt.Sprintf("**Overall Status:** %s\n\n", result.ProjectStatus.Status))
	
	if len(result.ProjectStatus.Conditions) > 0 {
		output.WriteString("### Conditions\n\n")
		for _, cond := range result.ProjectStatus.Conditions {
			statusIcon := "✅"
			if cond.Status == "ERROR" {
				statusIcon = "❌"
			} else if cond.Status == "WARNING" {
				statusIcon = "⚠️"
			}
			output.WriteString(fmt.Sprintf("%s **%s**\n", statusIcon, cond.MetricKey))
			output.WriteString(fmt.Sprintf("   - Value: %s\n", cond.MeasureValue))
			if cond.WarningThreshold != "" {
				output.WriteString(fmt.Sprintf("   - Warning: %s\n", cond.WarningThreshold))

			if cond.ErrorThreshold != "" {
				output.WriteString(fmt.Sprintf("   - Error: %s\n", cond.ErrorThreshold))

			output.WriteString("\n")

	}

	return ok(output.String())
}

}
}
}

// HandleSearchSonarIssuesInProjects searches for issues in SonarQube projects
func HandleSearchSonarIssuesInProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := newSonarQubeClient()
	
	if client.token == "" {
		return err("SONARQUBE_TOKEN environment variable is required")
}

	projectKey, _ :=getString(args, "projectKey")
	severities, _ :=getString(args, "severities")
	types, _ :=getString(args, "types")
	statuses, _ :=getString(args, "statuses")
	branch, _ :=getString(args, "branch")
	pullRequest, _ :=getString(args, "pullRequest")
	pageSize, _ :=getInt(args, "pageSize")
	if pageSize == 0 {
		pageSize = 100
	}
	page, _ :=getInt(args, "page")
	if page == 0 {
		page = 1
	}

	params := url.Values{}
	params.Set("ps", strconv.Itoa(pageSize))
	params.Set("p", strconv.Itoa(page))
	
	if projectKey != "" {
		params.Set("projects", projectKey)

	if severities != "" {
		params.Set("severities", severities)

	if types != "" {
		params.Set("types", types)

	if statuses != "" {
		params.Set("statuses", statuses)

	if branch != "" {
		params.Set("branch", branch)

	if pullRequest != "" {
		params.Set("pullRequest", pullRequest)

	response, apiErr := client.doRequest(ctx, "GET", "/api/issues/search", params, nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result struct {
		Total int `json:"total"`
		Paging struct {
			PageIndex int `json:"pageIndex"`
			PageSize  int `json:"pageSize"`
		} `json:"paging"`
		Issues []struct {
			Key          string `json:"key"`
			Type         string `json:"type"`
			Severity     string `json:"severity"`
			Message      string `json:"message"`
			Line         int    `json:"line"`
			Component    string `json:"component"`
			Project      string `json:"project"`
			Status       string `json:"status"`
			Resolution   string `json:"resolution"`
			Rule         string `json:"rule"`
			Tags         []string `json:"tags"`
		} `json:"issues"`
	}

	parseErr := json.Unmarshal(response, &result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d issues (page %d):\n\n", result.Total, result.Paging.PageIndex))
	
	for _, issue := range result.Issues {
		severityIcon := "ℹ️"
		switch issue.Severity {
		case "BLOCKER":
			severityIcon = "🔴"
		case "CRITICAL":
			severityIcon = "🟠"
		case "MAJOR":
			severityIcon = "🟡"
		case "MINOR":
			severityIcon = "🔵"
		}
		
		output.WriteString(fmt.Sprintf("### %s %s [%s]\n", severityIcon, issue.Type, issue.Severity))
		output.WriteString(fmt.Sprintf("**Key:** `%s`\n", issue.Key))
		output.WriteString(fmt.Sprintf("**Message:** %s\n", issue.Message))
		output.WriteString(fmt.Sprintf("**Component:** `%s`\n", issue.Component))
		if issue.Line > 0 {
			output.WriteString(fmt.Sprintf("**Line:** %d\n", issue.Line))

		output.WriteString(fmt.Sprintf("**Status:** %s\n", issue.Status))
		if issue.Resolution != "" {
			output.WriteString(fmt.Sprintf("**Resolution:** %s\n", issue.Resolution))

		output.WriteString(fmt.Sprintf("**Rule:** %s\n", issue.Rule))
		if len(issue.Tags) > 0 {
			output.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(issue.Tags, ", ")))

		output.WriteString("\n")

	return ok(output.String())
}

}
}
}
}
}
}
}
}
}
}

// HandleAnalyzeCodeSnippet analyzes a code snippet for issues
func HandleAnalyzeCodeSnippet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := newSonarQubeClient()
	
	if client.token == "" {
		return err("SONARQUBE_TOKEN environment variable is required")
}

	fileContent, _ :=getString(args, "fileContent")
	codeSnippet, _ :=getString(args, "codeSnippet")
	language, _ :=getString(args, "language")
	
	if fileContent == "" {
		return err("fileContent is required")
}

	params := url.Values{}
	params.Set("k", "sonarqube-mcp-snippet-analysis")
	
	requestBody := map[string]interface{}{
		"content":   fileContent,
		"language":  language,
	}
	
	if codeSnippet != "" {
		requestBody["snippet"] = codeSnippet
	}

	response, apiErr := client.doRequest(ctx, "POST", "/api/issues/sarif", params, requestBody)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result struct {
		Sarif string `json:"sarif"`
	}
	
	parseErr := json.Unmarshal(response, &result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	var output strings.Builder
	output.WriteString("## Code Snippet Analysis Results\n\n")
	
	if result.Sarif == "" {
		output.WriteString("No issues found in the provided code snippet.\n")
	} else {
		// Parse SARIF for summary
		var sarif struct {
			Runs []struct {
				Results []struct {
					RuleID    string `json:"ruleId"`
					Level     string `json:"level"`
					Message   struct {
						Text string `json:"text"`
					} `json:"message"`
					Locations []struct {
						PhysicalLocation struct {
							Region struct {
								StartLine int `json:"startLine"`
								StartColumn int `json:"startColumn"`
							} `json:"region"`
						} `json:"physicalLocation"`
					} `json:"locations"`
				} `json:"results"`
			} `json:"runs"`
		}
		
		if jsonErr := json.Unmarshal([]byte(result.Sarif), &sarif); jsonErr == nil {
			issueCount := 0
			for _, run := range sarif.Runs {
				issueCount += len(run.Results)

			output.WriteString(fmt.Sprintf("Found %d issue(s):\n\n", issueCount))
			
			for _, run := range sarif.Runs {
				for _, res := range run.Results {
					levelIcon := "ℹ️"
					if res.Level == "error" {
						levelIcon = "❌"
					} else if res.Level == "warning" {
						levelIcon = "⚠️"
					}
					
					output.WriteString(fmt.Sprintf("%s **%s**\n", levelIcon, res.RuleID))
					output.WriteString(fmt.Sprintf("   %s\n", res.Message.Text))
					
					if len(res.Locations) > 0 {
						loc := res.Locations[0].PhysicalLocation.Region
						if loc.StartLine > 0 {
							output.WriteString(fmt.Sprintf("   Location: Line %d", loc.StartLine))
							if loc.StartColumn > 0 {
								output.WriteString(fmt.Sprintf(", Column %d", loc.StartColumn))

							output.WriteString("\n")

					}
					output.WriteString("\n")

			}
		} else {
			output.WriteString("Analysis complete. See SARIF output for details.\n")

	}

	return ok(output.String())
}

}
}
}
}
}

// HandleShowRule displays details about a specific SonarQube rule
func HandleShowRule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := newSonarQubeClient()
	
	if client.token == "" {
		return err("SONARQUBE_TOKEN environment variable is required")
}

	ruleKey, _ :=getString(args, "ruleKey")
	if ruleKey == "" {
		return err("ruleKey is required")
}

	params := url.Values{}
	params.Set("key", ruleKey)

	response, apiErr := client.doRequest(ctx, "GET", "/api/rules/show", params, nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result struct {
	}
}