package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// LeadUpdate represents a change in lead status from the CRM.
type LeadUpdate struct {
	ID        string
	NewState  db.LeadState
	Notes     string
}

// CRMClient defines the interface for interacting with external CRM systems.
type CRMClient interface {
	// PushDeal synchronizes a local deal to the CRM.
	PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error

	// GetLeadUpdates retrieves status changes from the CRM for local reconciliation.
	GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error)

	// ValidateAccount checks if a company account is valid and active in the CRM.
	ValidateAccount(ctx context.Context, domain string) (bool, error)

	// SyncInteraction pushes a specific interaction or note to the CRM deal.
	SyncInteraction(ctx context.Context, dealID int64, note string) error

	// SyncContacts synchronizes contacts for a specific company to the CRM.
	SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error

	// FetchDealDetails retrieves specific deal information from the CRM.
	FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error)
}

// DealDetails represents detailed information for a deal in the CRM.
type DealDetails struct {
	ID                 int64        `json:"id"`
	Status             db.LeadState `json:"status"`
	QuotedPricing      float64      `json:"quoted_pricing"`
	CustomRequirements string       `json:"custom_requirements"`
	TechnicalDossier   string       `json:"technical_dossier"`
}

// RestCRMClient implements CRMClient using a generic REST API.
type RestCRMClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewRestCRMClient creates a new REST-based CRM client.
func NewRestCRMClient(baseURL, apiKey string) *RestCRMClient {
	return &RestCRMClient{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *RestCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	url := fmt.Sprintf("%s/deals", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"deal_id":           deal.ID,
		"company":           company.Name,
		"status":            deal.CurrentState,
		"pricing":           deal.QuotedPricing,
		"technical_dossier": deal.TechnicalDossier,
		"route":             route,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("crm api error: %d", resp.StatusCode)
	}

	return nil
}

func (c *RestCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	url := fmt.Sprintf("%s/companies/%d/contacts", c.BaseURL, companyID)
	payload, _ := json.Marshal(contacts)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("crm api error: %d", resp.StatusCode)
	}

	return nil
}

func (c *RestCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	url := fmt.Sprintf("%s/updates", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("crm api error: %d", resp.StatusCode)
	}

	var updates []LeadUpdate
	if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
		return nil, err
	}

	return updates, nil
}

func (c *RestCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	url := fmt.Sprintf("%s/accounts/validate?domain=%s", c.BaseURL, domain)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("crm api error: %d", resp.StatusCode)
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (c *RestCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/deals/%d", c.BaseURL, dealID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crm api error: %d", resp.StatusCode)
	}

	var details DealDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

func (c *RestCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	url := fmt.Sprintf("%s/deals/%d/interactions", c.BaseURL, dealID)
	payload, _ := json.Marshal(map[string]interface{}{
		"note": note,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("crm api error: %d", resp.StatusCode)
	}

	return nil
}
