package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleGetLatestVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	group, _ :=getString(args, "group")
	artifact, _ :=getString(args, "artifact")
	if group == "" || artifact == "" {
		return err("group and artifact are required")
}

	query := url.Values{}
	query.Set("q", fmt.Sprintf(`g:"%s" AND a:"%s"`, group, artifact))
	query.Set("rows", "1")
	query.Set("wt", "json")
	resp, e := http.DefaultClient.Get("https://search.maven.org/solrsearch/select?" + query.Encode())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result struct {
		Response struct {
			Docs []struct {
				LatestVersion string `json:"latestVersion"`
			} `json:"docs"`
		} `json:"response"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	if len(result.Response.Docs) == 0 {
		return err("artifact not found")
}

	return ok(fmt.Sprintf("Latest version: %s", result.Response.Docs[0].LatestVersion))
}

func HandleCheckVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	group, _ :=getString(args, "group")
	artifact, _ :=getString(args, "artifact")
	version, _ :=getString(args, "version")
	if group == "" || artifact == "" || version == "" {
		return err("group, artifact, and version are required")
}

	query := url.Values{}
	query.Set("q", fmt.Sprintf(`g:"%s" AND a:"%s" AND v:"%s"`, group, artifact, version))
	query.Set("rows", "1")
	query.Set("wt", "json")
	resp, e := http.DefaultClient.Get("https://search.maven.org/solrsearch/select?" + query.Encode())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if strings.Contains(string(body), "\"numFound\":0") {
		return ok(fmt.Sprintf("Version %s not found in Maven Central", version))
}

	return ok(fmt.Sprintf("Version %s exists in Maven Central", version))
}