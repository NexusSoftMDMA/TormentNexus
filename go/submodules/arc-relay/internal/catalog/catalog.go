package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const registryBaseURL = "https://registry.modelcontextprotocol.io/v0/servers"

// RegistryResponse is the top-level response from the MCP registry API.
type RegistryResponse struct {
	Servers  []RegistryEntry  `json:"servers"`
	Metadata RegistryMetadata `json:"metadata"`
}

// RegistryMetadata holds result counts.
type RegistryMetadata struct {
	Count int `json:"count"`
}

// RegistryEntry wraps a single registry server entry.
type RegistryEntry struct {
	Server ServerInfo `json:"server"`
}

// ServerInfo is the server metadata from the registry.
type ServerInfo struct {
	Name        string     `json:"name"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	Repository  Repository `json:"repository"`
	Packages    []Package  `json:"packages,omitempty"`
	Remotes     []Remote   `json:"remotes,omitempty"`
}

// Repository holds source repository info.
type Repository struct {
	URL    string `json:"url"`
	Source string `json:"source"`
}

// Package describes a distributable package (npm, pypi, oci).
type Package struct {
	RegistryType         string                `json:"registryType"`
	Identifier           string                `json:"identifier"`
	Version              string                `json:"version"`
	Transport            Transport             `json:"transport"`
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables,omitempty"`
}

// Transport describes how a package communicates.
type Transport struct {
	Type string `json:"type"`
}

// EnvironmentVariable describes a required/optional env var.
type EnvironmentVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsRequired  bool   `json:"isRequired"`
	IsSecret    bool   `json:"isSecret"`
	Format      string `json:"format"`
}

// Remote describes a remote server endpoint.
type Remote struct {
	Type    string   `json:"type"`
	URL     string   `json:"url"`
	Headers []Header `json:"headers,omitempty"`
}

// Header describes an HTTP header for remote connections.
type Header struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsRequired  bool   `json:"isRequired"`
	IsSecret    bool   `json:"isSecret"`
}

// ResolvedServer is what the catalog produces for the frontend — one per deployment option.
type ResolvedServer struct {
	RegistryName     string                `json:"registry_name"`
	Title            string                `json:"title"`
	Description      string                `json:"description"`
	Version          string                `json:"version"`
	RepoURL          string                `json:"repo_url"`
	SuggestedSlug    string                `json:"suggested_slug"`
	SuggestedDisplay string                `json:"suggested_display"`
	ServerType       string                `json:"server_type"`            // "stdio" or "remote"
	DockerImage      string                `json:"docker_image"`           // for OCI packages
	PackageType      string                `json:"package_type,omitempty"` // "npm", "pypi" — for auto-build
	PackageName      string                `json:"package_name,omitempty"` // package identifier — for auto-build
	EnvVars          []EnvironmentVariable `json:"env_vars,omitempty"`
	RemoteURL        string                `json:"remote_url"`  // for remotes
	RemoteType       string                `json:"remote_type"` // "sse" or "streamable-http"
}

// Client queries the MCP registry API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new registry client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Search queries the registry and returns relevance-ranked resolved servers.
// The registry only does substring matching on server names, so we over-fetch
// and re-rank client-side using slug, title, and description matching.
func (c *Client) Search(ctx context.Context, query string, limit int) ([]ResolvedServer, error) {
	if limit <= 0 {
		limit = 20
	}
	normalized := strings.ToLower(strings.TrimSpace(query))

	// The registry search is substring on names only. Normalize spaces to hyphens
	// since registry names use hyphens. Over-fetch to get enough candidates for scoring.
	registryQuery := strings.ReplaceAll(normalized, " ", "-")
	fetchLimit := 100

	resp, err := c.fetchRegistry(ctx, registryQuery, fetchLimit)
	if err != nil {
		return nil, err
	}

	// Resolve all entries and score them
	type scored struct {
		server ResolvedServer
		score  int
	}
	var candidates []scored

	for _, entry := range resp.Servers {
		resolved := Resolve(entry)
		s := relevanceScore(entry.Server, normalized)
		for _, rs := range resolved {
			candidates = append(candidates, scored{server: rs, score: s})
		}
	}

	// Sort by score descending, then by title
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].server.Title < candidates[j].server.Title
	})

	// Filter out irrelevant results (score 0 = only matched because of io.github.* prefix noise)
	var results []ResolvedServer
	for _, c := range candidates {
		if c.score <= 0 {
			continue
		}
		results = append(results, c.server)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// relevanceScore computes how well a registry entry matches the user's query.
// Higher = better. 0 = no real match (likely io.github.* noise).
func relevanceScore(info ServerInfo, query string) int {
	slug := deriveSlug(info.Name)
	queryHyphen := strings.ReplaceAll(query, " ", "-")
	nameLower := strings.ToLower(info.Name)
	descLower := strings.ToLower(info.Description)
	titleLower := strings.ToLower(info.Title)

	score := 0

	// Exact slug match: "sentry-mcp" searching "sentry"
	switch slug {
	case queryHyphen:
		score += 100
	case queryHyphen + "-mcp", "mcp-" + queryHyphen:
		score += 90 // common pattern: sentry-mcp, mcp-github
	}

	// Slug contains query
	if strings.Contains(slug, queryHyphen) {
		score += 50
	}

	// Title match (some entries have an explicit title field)
	if titleLower != "" && strings.Contains(titleLower, query) {
		score += 40
	}

	// Description contains query words
	queryWords := strings.Fields(query)
	descMatches := 0
	for _, w := range queryWords {
		if strings.Contains(descLower, w) {
			descMatches++
		}
	}
	if descMatches == len(queryWords) && len(queryWords) > 0 {
		score += 30
	} else if descMatches > 0 {
		score += 15
	}

	// The part after the last / in the name (the actual server name portion)
	namePart := nameLower
	if idx := strings.LastIndex(namePart, "/"); idx >= 0 {
		namePart = namePart[idx+1:]
	}
	if strings.Contains(namePart, queryHyphen) {
		score += 20
	}

	// Penalize if the only reason it matched is the io.github.* prefix
	// (the registry returns these for any "github"-adjacent query)
	if score == 0 {
		prefix := nameLower
		if idx := strings.LastIndex(prefix, "/"); idx >= 0 {
			prefix = prefix[:idx]
		}
		if strings.Contains(prefix, queryHyphen) && !strings.Contains(namePart, queryHyphen) && !strings.Contains(descLower, query) {
			// Only matched in the org/prefix portion, not the actual server name or description
			score = 0
		}
	}

	return score
}

func (c *Client) fetchRegistry(ctx context.Context, query string, limit int) (*RegistryResponse, error) {
	u := fmt.Sprintf("%s?search=%s&limit=%d", registryBaseURL, url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating registry request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading registry response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned %d: %s", resp.StatusCode, string(body))
	}

	var result RegistryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing registry response: %w", err)
	}
	return &result, nil
}

// Resolve converts a registry entry into one or more resolved servers (one per deployment option).
func Resolve(entry RegistryEntry) []ResolvedServer {
	info := entry.Server
	slug := deriveSlug(info.Name)
	display := deriveDisplay(info, slug)

	var results []ResolvedServer

	// Each remote → a remote server option
	for _, r := range info.Remotes {
		rs := ResolvedServer{
			RegistryName:     info.Name,
			Title:            display,
			Description:      info.Description,
			Version:          info.Version,
			RepoURL:          info.Repository.URL,
			SuggestedSlug:    slug,
			SuggestedDisplay: display,
			ServerType:       "remote",
			RemoteURL:        r.URL,
			RemoteType:       r.Type,
		}
		results = append(results, rs)
	}

	// OCI packages → stdio server with Docker image
	for _, pkg := range info.Packages {
		if pkg.RegistryType == "oci" {
			rs := ResolvedServer{
				RegistryName:     info.Name,
				Title:            display,
				Description:      info.Description,
				Version:          info.Version,
				RepoURL:          info.Repository.URL,
				SuggestedSlug:    slug,
				SuggestedDisplay: display,
				ServerType:       "stdio",
				DockerImage:      pkg.Identifier,
				EnvVars:          pkg.EnvironmentVariables,
			}
			results = append(results, rs)
		}
	}

	// npm/pypi packages — include as stdio hint with package metadata for auto-build
	if len(results) == 0 {
		for _, pkg := range info.Packages {
			if pkg.RegistryType == "npm" || pkg.RegistryType == "pypi" {
				rs := ResolvedServer{
					RegistryName:     info.Name,
					Title:            display,
					Description:      info.Description,
					Version:          info.Version,
					RepoURL:          info.Repository.URL,
					SuggestedSlug:    slug,
					SuggestedDisplay: display,
					ServerType:       "stdio",
					DockerImage:      "",
					PackageType:      pkg.RegistryType,
					PackageName:      pkg.Identifier,
					EnvVars:          pkg.EnvironmentVariables,
				}
				results = append(results, rs)
				break
			}
		}
	}

	return results
}

// deriveSlug extracts a URL-safe slug from a registry name like "io.github.getsentry/sentry-mcp".
func deriveSlug(name string) string {
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	slug := strings.ToLower(name)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, slug)
	slug = strings.Trim(slug, "-")
	return slug
}

// deriveDisplay creates a human-readable display name from the server info.
func deriveDisplay(info ServerInfo, slug string) string {
	// Prefer explicit title if available
	if info.Title != "" {
		return info.Title
	}
	words := strings.Split(slug, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
