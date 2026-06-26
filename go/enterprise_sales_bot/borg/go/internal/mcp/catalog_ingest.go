package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IngestResult struct {
	Source   string   `json:"source"`
	Fetched  int      `json:"fetched"`
	Upserted int      `json:"upserted"`
	Errors   []string `json:"errors"`
}

type IngestionReport struct {
	StartedAt     string         `json:"started_at"`
	FinishedAt    string         `json:"finished_at"`
	Results       []IngestResult `json:"results"`
	TotalUpserted int            `json:"total_upserted"`
	TotalErrors   int            `json:"total_errors"`
}

type CatalogSourceAdapter interface {
	Name() string
	Ingest(ctx context.Context, db *sql.DB) (IngestResult, error)
}

type BaseAdapter struct {
	name    string
	baseUrl string
}

func (a *BaseAdapter) Name() string {
	return a.name
}

func safeFetchJSON(ctx context.Context, client *http.Client, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TormentNexus/MCP-Catalog-Ingestor-Go")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// ---------------------------------------------------------------------------
// Helpers ported from TS
// ---------------------------------------------------------------------------

func InferRequiredSecrets(displayName, description string, tags, categories []string, authModel string) []string {
	authModel = strings.ToLower(authModel)
	if authModel == "none" || authModel == "public" {
		return []string{}
	}

	candidates := make(map[string]bool)
	scanTargets := strings.Join(append([]string{description}, append(tags, categories...)...), " ")

	// Pattern: UPPER_SNAKE_CASE env var name ending in a well-known secret suffix
	pattern := regexp.MustCompile(`\b([A-Z][A-Z0-9_]{1,}(?:_API_KEY|_TOKEN|_SECRET|_PASSWORD|_AUTH_TOKEN|_BEARER_TOKEN|_ACCESS_TOKEN|_ACCESS_KEY|_CLIENT_SECRET|_API_TOKEN|_AUTH_KEY|_PRIVATE_KEY|_SECRET_KEY))\b`)
	matches := pattern.FindAllStringSubmatch(scanTargets, -1)
	for _, m := range matches {
		if len(m[1]) <= 48 {
			candidates[m[1]] = true
		}
	}

	knownProviders := []struct {
		pattern *regexp.Regexp
		envVar  string
	}{
		{regexp.MustCompile(`(?i)github`), "GITHUB_TOKEN"},
		{regexp.MustCompile(`(?i)openai`), "OPENAI_API_KEY"},
		{regexp.MustCompile(`(?i)anthropic`), "ANTHROPIC_API_KEY"},
		{regexp.MustCompile(`(?i)google|gemini`), "GOOGLE_API_KEY"},
		{regexp.MustCompile(`(?i)slack`), "SLACK_BOT_TOKEN"},
		{regexp.MustCompile(`(?i)stripe`), "STRIPE_SECRET_KEY"},
		{regexp.MustCompile(`(?i)notion`), "NOTION_API_KEY"},
		{regexp.MustCompile(`(?i)jira|atlassian`), "JIRA_API_TOKEN"},
		{regexp.MustCompile(`(?i)linear`), "LINEAR_API_KEY"},
		{regexp.MustCompile(`(?i)discord`), "DISCORD_BOT_TOKEN"},
	}

	if len(candidates) == 0 && authModel != "unknown" {
		fullText := displayName + " " + scanTargets
		for _, p := range knownProviders {
			if p.pattern.MatchString(fullText) {
				candidates[p.envVar] = true
				break
			}
		}
	}

	if len(candidates) == 0 && (authModel == "api_key" || authModel == "bearer" || authModel == "token") {
		reg := regexp.MustCompile(`[^A-Z0-9]+`)
		safeName := reg.ReplaceAllString(strings.ToUpper(displayName), "_")
		safeName = strings.Trim(safeName, "_")
		if len(safeName) > 24 {
			safeName = safeName[:24]
		}
		if len(safeName) >= 2 {
			candidates[safeName+"_API_KEY"] = true
		}
	}

	result := []string{}
	for c := range candidates {
		result = append(result, c)
		if len(result) >= 4 {
			break
		}
	}
	return result
}

func BuildBaselineRecipe(displayName, transport, installMethod, repositoryUrl string, npmVersion *string) (map[string]any, int, string) {
	// Simplified port of baseline recipe logic
	// In a real implementation, we would want full parity with TS buildBaselineRecipe
	recipe := map[string]any{
		"type":    "stdio",
		"command": "npx",
		"args":    []string{"-y", strings.ToLower(strings.ReplaceAll(displayName, " ", "-"))},
		"env":     map[string]string{},
	}

	confidence := 20
	explanation := "Baseline Configurator fallback recipe generated from catalog metadata."

	return recipe, confidence, explanation
}

// ---------------------------------------------------------------------------
// Glama.ai Adapter
// ---------------------------------------------------------------------------

type GlamaServer struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Repository  struct {
		URL   string `json:"url"`
		Owner string `json:"owner"`
		Name  string `json:"name"`
	} `json:"repository"`
	Vendor struct {
		Name string `json:"name"`
	} `json:"vendor"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	Attributes struct {
		Stars int `json:"stars"`
	} `json:"attributes"`
	Transport string `json:"transport"`
}

type GlamaAiAdapter struct {
	BaseAdapter
}

func NewGlamaAiAdapter() *GlamaAiAdapter {
	return &GlamaAiAdapter{
		BaseAdapter: BaseAdapter{
			name:    "glama.ai",
			baseUrl: "https://glama.ai/api/mcp/servers",
		},
	}
}

func (a *GlamaAiAdapter) Ingest(ctx context.Context, db *sql.DB) (IngestResult, error) {
	result := IngestResult{
		Source: a.name,
		Errors: []string{},
	}

	client := &http.Client{Timeout: 15 * time.Second}
	var payload struct {
		Servers []GlamaServer `json:"servers"`
	}

	err := safeFetchJSON(ctx, client, a.baseUrl+"?limit=200", &payload)
	if err != nil {
		return result, err
	}

	result.Fetched = len(payload.Servers)

	for _, s := range payload.Servers {
		canonicalId := a.buildCanonicalId(s)

		// Upsert logic here
		// We'll need a shared repository or just execute the query
		serverUuid, err := upsertPublishedCatalogServer(db, map[string]any{
			"canonical_id":   canonicalId,
			"display_name":   s.Name,
			"description":    s.Description,
			"author":         s.Vendor.Name,
			"repository_url": s.Repository.URL,
			"tags":           append(s.Tags, s.Categories...),
			"categories":     s.Categories,
			"stars":          s.Attributes.Stars,
			"transport":      a.normalizeTransport(s.Transport),
			"install_method": "unknown",
			"auth_model":     "unknown",
		})

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("glama server %s: %v", s.ID, err))
			continue
		}

		err = upsertPublishedCatalogSource(db, serverUuid, a.name, a.baseUrl+"/"+s.Slug, s)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("glama source %s: %v", s.ID, err))
			continue
		}

		result.Upserted++
	}

	return result, nil
}

func (a *GlamaAiAdapter) buildCanonicalId(s GlamaServer) string {
	if s.Repository.Owner != "" && s.Repository.Name != "" {
		return fmt.Sprintf("github/%s/%s", s.Repository.Owner, s.Repository.Name)
	}
	if s.Slug != "" {
		return fmt.Sprintf("glama/%s", s.Slug)
	}
	return fmt.Sprintf("glama/%s", strings.ToLower(strings.ReplaceAll(s.Name, " ", "-")))
}

func (a *GlamaAiAdapter) normalizeTransport(raw string) string {
	t := strings.ToLower(raw)
	if strings.Contains(t, "stdio") {
		return "stdio"
	}
	if strings.Contains(t, "sse") {
		return "sse"
	}
	if strings.Contains(t, "http") {
		return "streamable_http"
	}
	return "unknown"
}

// ---------------------------------------------------------------------------
// DB Helpers
// ---------------------------------------------------------------------------

func upsertPublishedCatalogServer(db *sql.DB, data map[string]any) (string, error) {
	// Check if already exists by canonical_id
	var uuid string
	err := db.QueryRow("SELECT uuid FROM published_mcp_servers WHERE canonical_id = ?", data["canonical_id"]).Scan(&uuid)

	now := time.Now().UTC().Format(time.RFC3339)

	if err == sql.ErrNoRows {
		uuid = createUUID()
		_, err = db.Exec(`
			INSERT INTO published_mcp_servers (
				uuid, canonical_id, display_name, description, author, 
				repository_url, homepage_url, icon_url, transport, 
				install_method, auth_model, stars, status, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			uuid, data["canonical_id"], data["display_name"], data["description"], data["author"],
			data["repository_url"], data["homepage_url"], data["icon_url"], data["transport"],
			data["install_method"], data["auth_model"], data["stars"], "discovered", now, now,
		)
		return uuid, err
	} else if err != nil {
		return "", err
	}

	// Update existing
	_, err = db.Exec(`
		UPDATE published_mcp_servers SET 
			display_name = ?, description = ?, author = ?, 
			repository_url = ?, homepage_url = ?, icon_url = ?, transport = ?, 
			install_method = ?, auth_model = ?, stars = ?, updated_at = ?
		WHERE uuid = ?`,
		data["display_name"], data["description"], data["author"],
		data["repository_url"], data["homepage_url"], data["icon_url"], data["transport"],
		data["install_method"], data["auth_model"], data["stars"], now, uuid,
	)

	return uuid, err
}

func upsertPublishedCatalogSource(db *sql.DB, serverUuid, sourceName, sourceUrl string, rawPayload any) error {
	payloadJson, _ := json.Marshal(rawPayload)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.Exec(`
		INSERT INTO published_mcp_server_sources (
			server_uuid, source_name, source_url, raw_payload, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(server_uuid, source_name) DO UPDATE SET
			source_url = excluded.source_url,
			raw_payload = excluded.raw_payload,
			updated_at = excluded.updated_at`,
		serverUuid, sourceName, sourceUrl, string(payloadJson), now, now,
	)
	return err
}

func createUUID() string {
	return uuid.New().String()
}

func IngestPublishedCatalog(ctx context.Context, db *sql.DB) (IngestionReport, error) {
	startedAt := time.Now().UTC().Format(time.RFC3339)
	adapters := []CatalogSourceAdapter{
		NewGlamaAiAdapter(),
	}

	results := []IngestResult{}
	totalUpserted := 0
	totalErrors := 0

	for _, adapter := range adapters {
		res, err := adapter.Ingest(ctx, db)
		if err != nil {
			results = append(results, IngestResult{
				Source: adapter.Name(),
				Errors: []string{err.Error()},
			})
			totalErrors++
			continue
		}
		results = append(results, res)
		totalUpserted += res.Upserted
		totalErrors += len(res.Errors)
	}

	// Normalization and recipe passes would go here...

	return IngestionReport{
		StartedAt:     startedAt,
		FinishedAt:    time.Now().UTC().Format(time.RFC3339),
		Results:       results,
		TotalUpserted: totalUpserted,
		TotalErrors:   totalErrors,
	}, nil
}
