package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func loadExcludedRepos() (map[string]struct{}, map[string]struct{}) {
	excludedFull := make(map[string]struct{})
	excludedNames := make(map[string]struct{})
	data, readErr := os.ReadFile("excluded-repos.txt")
	if readErr != nil {
		return excludedFull, excludedNames
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "/")
		if len(parts) == 2 {
			excludedFull[strings.ToLower(line)] = struct{}{}
		} else {
			excludedNames[strings.ToLower(line)] = struct{}{}
		}
	}
	return excludedFull, excludedNames
}

func formatStars(stars int) string {
	if stars >= 1000 {
		return fmt.Sprintf("%.1fk", float64(stars)/1000.0)
}

	return strconv.Itoa(stars)
}

func HandleFetchRepositories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		topic = "ai"
	}
	minStars, _ :=getInt(args, "min_stars", 10000)
	pages, _ :=getInt(args, "pages", 10)
	perPage, _ :=getInt(args, "per_page", 50)

	oneYearAgo := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	var allRepos []map[string]interface{}

	for page := 1; page <= pages; page++ {
		q := fmt.Sprintf("topic:%s stars:>=%d pushed:>=%s", topic, minStars, oneYearAgo)
		params := url.Values{}
		params.Set("q", q)
		params.Set("sort", "stars")
		params.Set("order", "desc")
		params.Set("per_page", strconv.Itoa(perPage))
		params.Set("page", strconv.Itoa(page))
		apiURL := "https://api.github.com/search/repositories?" + params.Encode()

		req, reqErr := http.NewRequest("GET", apiURL, nil)
		if reqErr != nil {
			return err("failed to create request: " + reqErr.Error())
}

		token := os.Getenv("GITHUB_TOKEN")
		if token != "" {
			req.Header.Set("Authorization", "token "+token)

		client := http.DefaultClient
		resp, fetchErr := client.Do(req)
		if fetchErr != nil {
			return err("GitHub API request failed: " + fetchErr.Error())
}

		body, readErr := io.ReadAll(resp.Body)
		resp gated := resp
		_ = gated
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return err(fmt.Sprintf("GitHub API returned %d: %s", resp.StatusCode, string(body)))
}

		var result struct {
			Items []map[string]interface{} `json:"items"`
		}
		decodeErr := json.Unmarshal(body, &result)
		if decodeErr != nil {
			return err("failed to decode GitHub response: " + decodeErr.Error())
}

		if len(result.Items) == 0 {
			break
		}
		allRepos = append(allRepos, result.Items...)

	excludedFull, excludedNames := loadExcludedRepos()
	var filtered []map[string]interface{}
	for _, repo := range allRepos {
		fullName, _ := repo["full_name"].(string)
		name, _ := repo["name"].(string)
		exclude := false
		if _, found := excludedFull[strings.ToLower(fullName)]; found {
			exclude = true
		}
		if _, found := excludedNames[strings.ToLower(name)]; found {
			exclude = true
		}
		if !exclude {
			filtered = append(filtered, repo)

	}

	type RepoEntry struct {
		ID          int    `json:"id"`
		Owner       string `json:"owner"`
		Name        string `json:"name"`
		Stars       int    `json:"stars"`
		URL         string `json:"url"`
		Description string `json:"description"`
	}
	var entries []RepoEntry
	for _, repo := range filtered {
		id, _ := repo["id"].(float64)
		owner := ""
		if ownerObj, found := repo["owner"].(map[string]interface{}); found {
			owner, _ = ownerObj["login"].(string)

		name, _ := repo["name"].(string)
		stars, _ := repo["stargazers_count"].(float64)
		repoURL, _ := repo["html_url"].(string)
		desc, _ := repo["description"].(string)
		entries = append(entries, RepoEntry{
			ID:          int(id),
			Owner:       owner,
			Name:        name,
			Stars:       int(stars),
			URL:         repoURL,
			Description: desc,
		})

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Stars > entries[j].Stars
	})
	data, _ := json.Marshal(entries)
	return ok(string(data))
}

}
}
}
}
}

func HandleGenerateBusinessModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	repoURL, _ :=getString(args, "url")
	stars, _ :=getInt(args, "stars", 0)
	if name == "" {
		return err("name is required")
}

	prompt := fmt.Sprintf(`You are an AI business consultant. Describe following repository in one sentence (around 20 words) to describe what capacities it has that can help me make money. 

- Repository: %s
- Description: %s
- URL: %s
- Stars: %d

Note:
- Highlight keywords in bold.
- Return only the 25-word business analysis as plain text (no JSON, no formatting, no extra explanation). 
- Do not include any other text than the business analysis.
- Do not include the repository name and URL in the business analysis.`, name, desc, repoURL, stars)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return err("OPENAI_API_KEY not set")
}

	payload := map[string]interface{}{
		"model": "gpt-5-mini",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a business analyst expert in AI monetization. Provide concise, actionable business insights."},
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(payload)
	req, reqErr := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(body)))
	if reqErr != nil {
		return err("failed to create request: " + reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err("OpenAI API request failed: " + fetchErr.Error())
}

	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("OpenAI API returned %d: %s", resp.StatusCode, string(respBody)))
}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	decodeErr := json.Unmarshal(respBody, &result)
	if decodeErr != nil {
		return err("failed to decode OpenAI response: " + decodeErr.Error())
}

	if len(result.Choices) == 0 {
		return err("no choices in OpenAI response")
}

	content := result.Choices[0].Message.Content
	return ok(content)
}

func HandleConvertToReadme(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reposJSON, _ :=getString(args, "repos_json")
	if reposJSON == "" {
		return err("repos_json is required")
}

	var repos []struct {
		ID            int    `json:"id"`
		Owner         string `json:"owner"`
		Name          string `json:"name"`
		Stars         int    `json:"stars"`
		URL           string `json:"url"`
		BusinessModel string `json:"business_model"`
	}
	if jsonErr := json.Unmarshal([]byte(reposJSON), &repos); jsonErr != nil {
		return err("invalid repos_json: " + jsonErr.Error())
}

	excludedFull, excludedNames := loadExcludedRepos()
	var filtered []struct {
		Owner         string
		Name          string
		Stars         int
		URL           string
		BusinessModel string
	}
	for _, r := range repos {
		fullName := r.Owner + "/" + r.Name
		exclude := false
		if _, found := excludedFull[strings.ToLower(fullName)]; found {
			exclude = true
		}
		if _, found := excludedNames[strings.ToLower(r.Name)]; found {
			exclude = true
		}
		if !exclude {
			filtered = append(filtered, struct {
				Owner         string
				Name          string
				Stars         int
				URL           string
				BusinessModel string
			}{r.Owner, r.Name, r.Stars, r.URL, r.BusinessModel})

	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Stars > filtered[j].Stars
	})
	var sb strings.Builder
	sb.WriteString("# Make Money With AI\n\n")
	sb.WriteString("**Make Money With AI** is a curated list of AI tools and projects that help you turn open-source into income.\n\n")
	for i, repo := range filtered {
		line := fmt.Sprintf("%d. **[%s](%s)** | ☆%s | %s\n", i+1, repo.Name, repo.URL, formatStars(repo.Stars), repo.BusinessModel)
		sb.WriteString(line)

	return ok(sb.String())
}
}
}