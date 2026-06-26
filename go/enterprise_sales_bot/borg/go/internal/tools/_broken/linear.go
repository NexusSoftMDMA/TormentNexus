package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type linearIssue struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	Assignee    string `json:"assignee"`
	Team        string `json:"team"`
	CreatedAt   string `json:"createdAt"`
}

type linearTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type linearUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type linearGraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func linearRequest(apiKey string, query string, variables map[string]interface{}) ([]byte, error) {
	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, marshalErr
	}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(context.Background(), "POST", "https://api.linear.app/graphql", strings.NewReader(string(body)))
	if reqErr != nil {
		return nil, reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, doErr
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	return respBody, nil
}

func HandleLinearListIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	teamId, _ :=getString(args, "teamId")
	status, _ :=getString(args, "filterStatus")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	variables := map[string]interface{}{
		"teamId": teamId,
		"limit":  limit,
	}

	filter := ""
	if status != "" {
		filter = `, filter: { status: { eq: "${status}" } }`
		variables["status"] = status
	}

	query := `query($teamId: String!, $limit: Int!) {
		issues(filter: { team: { id: { eq: $teamId } }` + filter + `}, first: $limit) {
			nodes {
				id title description status { name } priority assignee { name } createdAt
			}
		}
	}`

	respBody, reqErr := linearRequest(apiKey, query, variables)
	if reqErr != nil {
		return err(reqErr.Error())
}

	var result linearGraphQLResponse
	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
		return err(parseErr.Error())
}

	if len(result.Errors) > 0 {
		return err(result.Errors[0].Message)
}

	var issues struct {
		Issues struct {
			Nodes []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Status      struct {
					Name string `json:"name"`
				} `json:"status"`
				Priority  int    `json:"priority"`
				Assignee  struct {
					Name string `json:"name"`
				} `json:"assignee"`
				CreatedAt string `json:"createdAt"`
			} `json:"nodes"`
		} `json:"issues"`
	}
	if unmarshalErr := json.Unmarshal(result.Data, &issues); unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	output, _ := json.MarshalIndent(issues.Issues.Nodes, "", "  ")
	return ok(string(output))
}

func HandleLinearCreateIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	teamId, _ :=getString(args, "teamId")
	title, _ :=getString(args, "title")
	description, _ :=getString(args, "description")
	priority, _ :=getInt(args, "priority")

	variables := map[string]interface{}{
		"teamId":      teamId,
		"title":       title,
		"description": description,
		"priority":    priority,
	}

	query := `mutation CreateIssue($teamId: String!, $title: String!, $description: String!, $priority: Int!) {
		issueCreate(input: { teamId: $teamId, title: $title, description: $description, priority: $priority }) {
			issue { id title status { name } }
			success
		}
	}`

	respBody, reqErr := linearRequest(apiKey, query, variables)
	if reqErr != nil {
		return err(reqErr.Error())
}

	var result linearGraphQLResponse
	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
		return err(parseErr.Error())
}

	if len(result.Errors) > 0 {
		return err(result.Errors[0].Message)
}

	var createResult struct {
		IssueCreate struct {
			Issue struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Status struct {
					Name string `json:"name"`
				} `json:"status"`
			} `json:"issue"`
			Success bool `json:"success"`
		} `json:"issueCreate"`
	}
	if unmarshalErr := json.Unmarshal(result.Data, &createResult); unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	output, _ := json.MarshalIndent(createResult.IssueCreate, "", "  ")
	return ok(string(output))
}

func HandleLinearListTeams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")

	query := `query { teams { nodes { id name key } } }`
	variables := map[string]interface{}{}

	respBody, reqErr := linearRequest(apiKey, query, variables)
	if reqErr != nil {
		return err(reqErr.Error())
}

	var result linearGraphQLResponse
	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
		return err(parseErr.Error())
}

	if len(result.Errors) > 0 {
		return err(result.Errors[0].Message)
}

	var teams struct {
		Teams struct {
			Nodes []linearTeam `json:"nodes"`
		} `json:"teams"`
	}
	if unmarshalErr := json.Unmarshal(result.Data, &teams); unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	output, _ := json.MarshalIndent(teams.Teams.Nodes, "", "  ")
	return ok(string(output))
}

func HandleLinearGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")

	query := `query { viewer { id name email } }`
	variables := map[string]interface{}{}

	respBody, reqErr := linearRequest(apiKey, query, variables)
	if reqErr != nil {
		return err(reqErr.Error())
}

	var result linearGraphQLResponse
	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
		return err(parseErr.Error())
}

	if len(result.Errors) > 0 {
		return err(result.Errors[0].Message)
}

	var viewer struct {
		Viewer linearUser `json:"viewer"`
	}
	if unmarshalErr := json.Unmarshal(result.Data, &viewer); unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	output, _ := json.MarshalIndent(viewer.Viewer, "", "  ")
	return ok(string(output))
}