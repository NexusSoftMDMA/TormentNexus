package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// HunterSource implements EnrichmentSource using the Hunter.io API
// to find professional email addresses for decision-makers at target companies.
//
// API docs: https://hunter.io/api-documentation
// Free tier: 25 searches/month, 50 verifications/month
type HunterSource struct {
	APIKey		string
	HTTPClient	*http.Client
}

// hunterResponse represents the Hunter.io email finder API response.

// hunterDomainResponse represents the Hunter.io domain search API response.
type hunterDomainResponse struct {
	Data	struct {
		Emails	[]struct {
			Value		string	`json:"value"`
			FirstName	string	`json:"first_name"`
			LastName	string	`json:"last_name"`
			Position	string	`json:"position"`
			Confidence	float64	`json:"confidence"`
			Seniority	string	`json:"seniority"`
			Department	string	`json:"department"`
			LinkedInURL	string	`json:"linkedin"`
		}	`json:"emails"`
		Pattern		string	`json:"pattern"`
		Organization	string	`json:"organization"`
	}	`json:"data"`
	Errors	[]string	`json:"errors"`
}

// NewHunterSource creates a new Hunter.io enrichment source.
func NewHunterSource(apiKey string) *HunterSource {
	return &HunterSource{
		APIKey:	apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Enrich implements EnrichmentSource. It uses Hunter.io's domain search
// to find decision-makers at the target company.
func (h *HunterSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	if h.APIKey == "" {
		return nil, fmt.Errorf("hunter: API key not configured")
	}

	// Clean domain — remove protocol, www, trailing slashes
	domain := cleanDomain(company.Domain)
	if domain == "" || strings.HasSuffix(domain, ".com") && len(domain) < 5 {
		return nil, fmt.Errorf("hunter: invalid domain %q", company.Domain)
	}

	slog.Info(fmt.Sprintf("HunterSource: Searching for contacts at %s (%s)", company.Name, domain))

	contacts, err := h.domainSearch(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("hunter: domain search failed for %s: %w", domain, err)
	}

	if len(contacts) == 0 {
		slog.Info(fmt.Sprintf("HunterSource: No contacts found at %s", domain))
		return nil, nil
	}

	// Filter for decision-makers (senior roles)
	var filtered []db.Contact
	for _, c := range contacts {
		if isDecisionMaker(c.Role) {
			filtered = append(filtered, c)
			slog.Info(fmt.Sprintf("HunterSource: Found decision-maker: %s (%s) at %s", c.Name, c.Role, domain))
		}
	}

	// If no decision-makers found, return top contacts anyway
	if len(filtered) == 0 && len(contacts) > 0 {
		filtered = contacts[:1]	// Take the first (highest confidence)
		slog.Info(fmt.Sprintf("HunterSource: No senior roles found, taking highest confidence contact: %s", filtered[0].Name))
	}

	return filtered, nil
}

// domainSearch calls Hunter.io's domain search endpoint.
func (h *HunterSource) domainSearch(ctx context.Context, domain string) ([]db.Contact, error) {
	apiURL := fmt.Sprintf("https://api.hunter.io/v2/domain-search?domain=%s&api_key=%s&limit=10&type=personal",
		url.QueryEscape(domain), url.QueryEscape(h.APIKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("hunter: invalid API key")
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("hunter: rate limited (free tier: 25 searches/month)")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hunter: API returned %d: %s", resp.StatusCode, string(body))
	}

	var result hunterDomainResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("hunter: failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("hunter: API error: %s", strings.Join(result.Errors, ", "))
	}

	var contacts []db.Contact
	for _, email := range result.Data.Emails {
		if email.Value == "" {
			continue
		}

		name := strings.TrimSpace(email.FirstName + " " + email.LastName)
		if name == "" {
			name = email.Value
		}

		role := email.Position
		if role == "" {
			role = fmt.Sprintf("%s Department", email.Department)
		}

		contacts = append(contacts, db.Contact{
			Name:		name,
			Role:		role,
			Email:		email.Value,
			LinkedInURL:	email.LinkedInURL,
		})
	}

	return contacts, nil
}

// HealthCheck verifies the Hunter.io API key is valid.
func (h *HunterSource) HealthCheck(ctx context.Context) error {
	if h.APIKey == "" {
		return fmt.Errorf("hunter: API key not configured")
	}

	apiURL := fmt.Sprintf("https://api.hunter.io/v2/account?api_key=%s", url.QueryEscape(h.APIKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return err
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("hunter: account check returned %d", resp.StatusCode)
	}

	return nil
}

// cleanDomain removes protocol, www, and trailing slashes from a domain.
func cleanDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.TrimSuffix(domain, "/")
	domain = strings.Split(domain, "/")[0]	// Remove path
	domain = strings.Split(domain, "?")[0]	// Remove query
	return strings.ToLower(domain)
}

// isDecisionMaker checks if a role/title indicates a senior decision-maker.
func isDecisionMaker(role string) bool {
	lower := strings.ToLower(role)

	seniorKeywords := []string{
		"cto", "vp", "vice president", "director", "head of", "chief",
		"principal", "staff", "lead", "manager", "architect",
		"founder", "co-founder", "ceo", "engineering manager",
		"team lead", "tech lead", "technical lead",
	}

	for _, kw := range seniorKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	return false
}
