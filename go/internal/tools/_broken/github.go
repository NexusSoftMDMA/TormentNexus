package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const githubAPIBase = "https://api.github.com"

func getGitHubToken() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")

	return token
}

}

func newGitHubRequest(ctx context.Context, method, apiURL string, body io.Reader) (*http.Request, error) {
	req, reqErr := http.NewRequestWithContext(ctx, method, apiURL, body)
	if reqErr != nil {
		return nil, reqErr
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	token := getGitHubToken()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	return req, nil
}

}

func doGitHubRequest(req *http.Request) ([]byte, error) {
	client := http.DefaultClient
	resp, reqErr := client.Do(req)
	if reqErr != nil {
		return nil, reqErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
}

	return body, nil
}

// HandleGitHubSearchRepositories searches for GitHub repositories
func HandleGitHubSearchRepositories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	v := url.Values{}
	v.Set("q", query)

	sortVal, _ :=getString(args, "sort")
	if sortVal != "" {
		v.Set("sort", sortVal)

	order, _ :=getString(args, "order")
	if order != "" {
		v.Set("order", order)

	perPage, _ :=getInt(args, "per_page")
	if perPage > 0 {
		v.Set("per_page", strconv.Itoa(perPage))
	} else {
		v.Set("per_page", "30")

	page, _ :=getInt(args, "page")
	if page > 0 {
		v.Set("page", strconv.Itoa(page))

	apiURL := githubAPIBase + "/search/repositories?" + v.Encode()
	req, reqErr := newGitHubRequest(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	body, doErr := doGitHubRequest(req)
	if doErr != nil {
		return err(doErr.Error())
}

	return ok(string(body))
}

}
}
}
}

// HandleGitHubGetRepository gets details of a specific repository
func HandleGitHubGetRepository(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo parameters are required")
}

	apiURL := fmt.Sprintf("%s/repos/%s/%s", githubAPIBase, owner, repo)
	req, reqErr := newGitHubRequest(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	body, doErr := doGitHubRequest(req)
	if doErr != nil {
		return err(doErr.Error())
}

	return ok(string(body))
}

// HandleGitHubListIssues lists issues in a repository
func HandleGitHubListIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo parameters are required")
}

	v := url.Values{}
	state, _ :=getString(args, "state")
	if state == "" {
		state = "open"
	}
	v.Set("state", state)

	labels, _ :=getString(args, "labels")
	if labels != "" {
		v.Set("labels", labels)

	assignee, _ :=getString(args, "assignee")
	if assignee != "" {
		v.Set("assignee", assignee)

	perPage, _ :=getInt(args, "per_page")
	if perPage > 0 {
		v.Set("per_page", strconv.Itoa(perPage))
	} else {
		v.Set("per_page", "30")

	page, _ :=getInt(args, "page")
	if page > 0 {
		v.Set("page", strconv.Itoa(page))

	apiURL := fmt.Sprintf("%s/repos/%s/%s/issues?%s", githubAPIBase, owner, repo, v.Encode())
	req, reqErr := newGitHubRequest(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	body, doErr := doGitHubRequest(req)
	if doErr != nil {
		return err(doErr.Error())
}

	return ok(string(body))
}

}
}
}
}

// HandleGitHubCreateIssue creates a new issue in a repository
func HandleGitHubCreateIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	title, _ :=getString(args, "title")
	if owner == "" || repo == "" || title == "" {
		return err("owner, repo, and title parameters are required")
}

	payload := map[string]interface{}{
		"title": title,
	}

	bodyText, _ :=getString(args, "body")
	if bodyText != "" {
		payload["body"] = bodyText
	}

	assignees, _ :=getString(args, "assignees")
	if assignees != "" {
		payload["assignees"] = []string{assignees}
	}

	milestone, _ :=getInt(args, "milestone")
	if milestone > 0 {
		payload["milestone"] = milestone
	}

	labels, _ :=getString(args, "labels")
	if labels != "" {
		payload["labels"] = []string{labels}
	}

	jsonBody, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	apiURL := fmt.Sprintf("%s/repos/%s/%s/issues", githubAPIBase, owner, repo)
	req, reqErr := newGitHubRequest(ctx, "POST", apiURL, strings.NewReader(string(jsonBody)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	body, doErr := doGitHubRequest(req)
	if doErr != nil {
		return err(doErr.Error())
}

	return ok(string(body))
}

// HandleGitHubGetFileContents gets the contents of a file from a repository
func HandleGitHubGetFileContents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	path, _ :=getString(args, "path")
	if owner == "" || repo == "" || path == "" {
		return err("owner, repo, and path parameters are required")
}

	v := url.Values{}
	ref, _ :=getString(args, "ref")
	if ref != "" {
		v.Set("ref", ref)

	apiURL := fmt.Sprintf("%s/repos/%s/%s/contents/%s", githubAPIBase, owner, repo, path)
	if len(v) > 0 {
		apiURL = apiURL + "?" + v.Encode()

	req, reqErr := newGitHubRequest(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	body, doErr := doGitHubRequest(req)
	if doErr != nil {
		return err(doErr.Error())
}

	return ok(string(body))
}

}
}

// HandleGitHubListPullRequests lists pull requests in a repository
func HandleGitHubListPullRequests(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo parameters are required")
}

	v := url.Values{}
	state, _ :=getString(args, "state")
	if state == "" {
		state = "open"
	}
	v.Set("state", state)

	head, _ :=getString(args, "head")
	if head != "" {
		v.Set("head", head)

	base, _ :=getString(args, "base")
	if base != "" {
		v.Set("base", base)

	sortVal, _ :=getString(args, "sort")
	if sortVal != "" {
		v.Set("sort", sortVal)

	direction, _ :=getString(args, "direction")
	if direction != "" {
		v.Set("direction", direction)

	perPage, _ :=getInt(args, "per_page")
	if perPage > 0 {
		v.Set("per_page", strconv.Itoa(perPage))
	} else {
		v.Set("per_page", "30")

	page, _ :=getInt(args, "page")
	if page > 0 {
		v.Set("page", strconv.Itoa(page))

	apiURL := fmt.Sprintf("%s/repos/%s/%s/pulls?%s", githubAPIBase, owner, repo, v.Encode())
	req, reqErr := newGitHubRequest(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	body, doErr := doGitHubRequest(req)
	if doErr != nil {
		return err(doErr.Error())
}

	return ok(string(body))
}
}
}
}
}
}
}