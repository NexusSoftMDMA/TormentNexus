package enrichment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// ApolloSource implements EnrichmentSource using the Apollo.io People Search API
// to find decision-makers at target companies.
//
// API docs: https://docs.apollo.io/reference/people-search
// Free tier: 5,000 credits/month (1 credit = 1 person returned)
type ApolloSource struct {
	APIKey		string
	HTTPClient	*http.Client
}

// apolloSearchRequest is the JSON body for the Apollo.io people search API.
type apolloSearchRequest struct {
	QOrganizationDomains	[]string	`json:"q_organization_domains"`
	PersonTitles		[]string	`json:"person_titles,omitempty"`
	Page			int		`json:"page"`
	PerPage			int		`json:"per_page"`
	RevealPhone		bool		`json:"reveal_phone"`
	RevealPersonalEmails	bool		`json:"reveal_personal_emails"`
	RequireEmail		bool		`json:"require_email"`
}

// apolloSearchResponse is the JSON response from the Apollo.io people search API.
type apolloSearchResponse struct {
	People		[]apolloPerson	`json:"people"`
	TotalEntries	int		`json:"total_entries"`
	Paginated	bool		`json:"paginated"`
}

// apolloPerson represents a single person record from the Apollo.io API.
type apolloPerson struct {
	ID		string	`json:"id"`
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	Name		string	`json:"name"`
	Title		string	`json:"title"`
	Email		string	`json:"email"`
	LinkedInURL	string	`json:"linkedin_url"`
	Organization	struct {
		Name string `json:"name"`
	}	`json:"organization"`
	Seniority	string	`json:"seniority"`
}

// NewApolloSource creates a new Apollo.io enrichment source.
func NewApolloSource(apiKey string) *ApolloSource {
	return &ApolloSource{
		APIKey:	apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Enrich implements EnrichmentSource. It uses Apollo.io's people search
// to find decision-makers at the target company by domain.
func (a *ApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	if a.APIKey == "" {
		return nil, fmt.Errorf("apollo: API key not configured")
	}

	domain := cleanDomain(company.Domain)
	if domain == "" {
		return nil, fmt.Errorf("apollo: invalid domain %q", company.Domain)
	}

	slog.Info(fmt.Sprintf("ApolloSource: Searching for contacts at %s (%s)", company.Name, domain))

	contacts, err := a.peopleSearch(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("apollo: people search failed for %s: %w", domain, err)
	}

	if len(contacts) == 0 {
		slog.Info(fmt.Sprintf("ApolloSource: No contacts found at %s", domain))
		return nil, nil
	}

	// Filter for decision-makers (senior titles)
	var filtered []db.Contact
	for _, c := range contacts {
		if isDecisionMaker(c.Role) {
			filtered = append(filtered, c)
			slog.Info(fmt.Sprintf("ApolloSource: Found decision-maker: %s (%s) at %s", c.Name, c.Role, domain))
		}
	}

	// If no decision-makers found, return top contacts anyway
	if len(filtered) == 0 && len(contacts) > 0 {
		filtered = contacts[:1]
		slog.Info(fmt.Sprintf("ApolloSource: No senior roles found, taking highest confidence contact: %s", filtered[0].Name))
	}

	return filtered, nil
}

// peopleSearch calls the Apollo.io people search API with filters for decision-maker titles.
func (a *ApolloSource) peopleSearch(ctx context.Context, domain string) ([]db.Contact, error) {
	reqBody := apolloSearchRequest{
		QOrganizationDomains:	[]string{domain},
		PersonTitles: []string{
			"CTO", "VP", "Vice President", "Director", "Chief",
			"Head of", "Principal", "Staff", "Lead", "Architect",
			"Engineering Manager", "Founder", "Co-Founder", "CEO",
		},
		Page:		1,
		PerPage:	25,
		RequireEmail:	true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.apollo.io/api/v1/people/search",
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Api-Key", a.APIKey)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("apollo: invalid API key (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("apollo: rate limited")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("apollo: API returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result apolloSearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var contacts []db.Contact
	for _, p := range result.People {
		if p.Email == "" {
			continue
		}

		name := strings.TrimSpace(p.Name)
		if name == "" {
			name = strings.TrimSpace(p.FirstName + " " + p.LastName)
		}
		if name == "" {
			name = p.Email
		}

		role := strings.TrimSpace(p.Title)
		if role == "" {
			role = fmt.Sprintf("Team Member (Seniority: %s)", p.Seniority)
		}

		contacts = append(contacts, db.Contact{
			Name:		name,
			Role:		role,
			Email:		p.Email,
			LinkedInURL:	p.LinkedInURL,
		})
	}

	return contacts, nil
}

// HealthCheck verifies the Apollo.io API key is valid by making a small search request.
func (a *ApolloSource) HealthCheck(ctx context.Context) error {
	if a.APIKey == "" {
		return fmt.Errorf("apollo: API key not configured")
	}

	reqBody := apolloSearchRequest{
		QOrganizationDomains:	[]string{"openai.com"},
		Page:			1,
		PerPage:		1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.apollo.io/api/v1/people/search",
		bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", a.APIKey)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("apollo health: connection failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("apollo health: API returned HTTP %d", resp.StatusCode)
	}

	return nil
}
