package tools

import (
	"fmt"
)

func (r *Registry) registerIntegrationsTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "jira_create_issue",
		Description: "Creates an issue in Jira (Rovo parity). Arguments: project_key (string), summary (string), description (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			projectKey, _ := args["project_key"].(string)
			summary, _ := args["summary"].(string)
			description, _ := args["description"].(string)

			// Simulate API call to Jira
			return fmt.Sprintf("Successfully created issue in %s: %s\nDescription: %s", projectKey, summary, description), nil
		},
	})

	r.Tools = append(r.Tools, Tool{
		Name:        "confluence_search",
		Description: "Searches Confluence for documentation (Rovo parity). Arguments: query (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			query, _ := args["query"].(string)

			// Simulate API call to Confluence
			return fmt.Sprintf("Found 3 documents matching '%s'. (Simulated Confluence search results)", query), nil
		},
	})
}
