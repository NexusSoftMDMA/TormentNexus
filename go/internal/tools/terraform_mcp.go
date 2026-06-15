//go:build ignore
// +build ignore

package tools

/**
 * @file terraform_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Terraform MCP — Terraform Registry and workspace management.
 * Replaces: github.com/hashicorp/terraform-mcp-server
 *
 * Features: Terraform Registry provider/module search, workspace operations,
 * variable management via HCP Terraform API.
 *
 * Tools:
 *  - terraform_search_providers — search Terraform Registry providers
 *  - terraform_search_modules — search Terraform Registry modules
 *  - terraform_get_provider — get provider details
 *  - terraform_list_workspaces — list HCP workspaces
 */

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const tfRegistryURL = "https://registry.terraform.io/v1"

type tfProvider struct {
	ID           string `json:"id"`
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	Source       string `json:"source"`
	Downloads    int    `json:"downloads"`
	Tier         string `json:"tier"`
}

type tfSearchResult struct {
	Providers []tfProvider `json:"providers"`
	Meta      struct {
		Limit      int `json:"limit"`
		Offset     int `json:"offset"`
		Total      int `json:"total"`
	} `json:"meta"`
}

// HandleTerraformSearchProviders searches providers in the Terraform Registry.
func HandleTerraformSearchProviders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	client := &http.Client{Timeout: 30 * time.Second}
	reqURL := fmt.Sprintf("%s/providers?q=%s&limit=%d",
		tfRegistryURL, url.QueryEscape(query), limit)
	resp, e := client.Get(reqURL)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleTerraformSearchModules searches modules in the Terraform Registry.
func HandleTerraformSearchModules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf("%s/modules?q=%s&limit=%d",
		tfRegistryURL, url.QueryEscape(query), limit))
	if e != nil {
		return err(fmt.Sprintf("search modules failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleTerraformGetProvider gets provider details from the Registry.
func HandleTerraformGetProvider(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ := getString(args, "namespace", "ns")
	name, _ := getString(args, "name", "provider")
	if namespace == "" || name == "" {
		return err("namespace and name are required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf("%s/providers/%s/%s",
		tfRegistryURL, url.PathEscape(namespace), url.PathEscape(name)))
	if e != nil {
		return err(fmt.Sprintf("get provider failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}
