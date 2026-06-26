package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleSearchClojars(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	u := fmt.Sprintf("https://clojars.org/api/search?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to search Clojars: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Results []struct {
			GroupID    string `json:"group_id"`
			ArtifactID string `json:"artifact_id"`
			Latest     string `json:"latest_release"`
			Description string `json:"description"`
		} `json:"results"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if len(result.Results) == 0 {
		return ok("No packages found.")
}

	var sb strings.Builder
	for _, r := range result.Results {
		sb.WriteString(fmt.Sprintf("%s/%s %s - %s\n", r.GroupID, r.ArtifactID, r.Latest, r.Description))

	return ok(sb.String())
}

}

func HandleGetPackageInfo_clojars_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	group, _ :=getString(args, "group")
	artifact, _ :=getString(args, "artifact")
	if group == "" || artifact == "" {
		return err("group and artifact parameters are required")
}

	u := fmt.Sprintf("https://clojars.org/api/artifacts/%s/%s", url.PathEscape(group), url.PathEscape(artifact))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch package info: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	latest, _ := data["latest_release"].(string)
	desc, _ := data["description"].(string)
	recent, _ := data["recent_versions"].([]interface{})
	var versions []string
	for _, v := range recent {
		if s, found := v.(string); found {
			versions = append(versions, s)

	}
	output := fmt.Sprintf("%s/%s\nLatest: %s\nDescription: %s\nRecent versions: %s", group, artifact, latest, desc, strings.Join(versions, ", "))
	return ok(output)
}
}