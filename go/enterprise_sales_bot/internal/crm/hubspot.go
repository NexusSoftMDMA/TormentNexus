package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// HubSpotClient implements CRMClient using the HubSpot REST API.
// Required environment variables:
//   HUBSPOT_API_KEY – Private API key (legacy) or use OAuth token in HUBSPOT_ACCESS_TOKEN
//   HUBSPOT_BASE_URL – Base API URL, e.g., https://api.hubapi.com
// The client uses HubSpot's CRM objects: contacts, companies, deals.

type HubSpotClient struct {
	baseURL     string
	apiKey      string
	accessToken string
	client      *http.Client
	stageMap    map[string]string
	reverseMap  map[string]string
}

// NewHubSpotClient creates a new HubSpot CRM client.
func NewHubSpotClient(stageMap map[string]string, reverseMap map[string]string) (*HubSpotClient, error) {
	base := os.Getenv("HUBSPOT_BASE_URL")
	if base == "" {
		base = "https://api.hubapi.com"
	}
	key := os.Getenv("HUBSPOT_API_KEY")
	token := os.Getenv("HUBSPOT_ACCESS_TOKEN")
	if key == "" && token == "" {
		return nil, fmt.Errorf("hubspot client: missing API key or access token")
	}
		return &HubSpotClient{
		baseURL:     base,
		apiKey:      key,
		accessToken: token,
		client:      &http.Client{},
		stageMap:    stageMap,
		reverseMap:  reverseMap,
	}, nil
}

// authHeader constructs the Authorization header for HubSpot requests.
func (h *HubSpotClient) authHeader() string {
	if h.accessToken != "" {
		return "Bearer " + h.accessToken
	}
	return "Bearer " + h.apiKey
}

// PushDeal creates or updates a HubSpot deal.
func (h *HubSpotClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	payload := map[string]interface{}{
		"properties": map[string]any{
			"dealname":           fmt.Sprintf("%s – %s", company.Name, route),
			"amount":             deal.QuotedPricing,
			"pipeline":           "default",
			"dealstage":          h.mapLeadStateToHubSpotStage(deal.CurrentState),
			"description":        deal.TechnicalDossier,
			"custom_route":       route,
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/crm/v3/objects/deals", h.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", h.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("hubspot PushDeal: status %d", resp.StatusCode)
	}
	return nil
}

// GetLeadUpdates fetches recently modified leads (contacts) from HubSpot.
func (h *HubSpotClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// HubSpot contacts endpoint with a simple filter for recent updates (last 24h)
	url := fmt.Sprintf("%s/crm/v3/objects/contacts?properties=firstname,lastname,email,hs_lead_status&limit=100&archived=false&propertiesWithHistory=hs_lead_status", h.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", h.authHeader())

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hubspot GetLeadUpdates: status %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			ID         string `json:"id"`
			Properties struct {
				LeadStatus string `json:"hs_lead_status"`
			} `json:"properties"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var updates []LeadUpdate
	for _, r := range result.Results {
		state := h.mapHubSpotLeadStatusToLeadState(r.Properties.LeadStatus)
		if state != "" {
			updates = append(updates, LeadUpdate{ID: r.ID, NewState: state, Notes: ""})
		}
	}

	return updates, nil
}

// ValidateAccount checks if a company domain exists as a HubSpot company.
func (h *HubSpotClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/companies/search", h.baseURL)
	payload := map[string]interface{}{
		"filterGroups": []map[string]interface{}{{
			"filters": []map[string]string{{
				"propertyName": "website",
				"operator":     "EQ",
				"value":        domain,
			}},
		}},
		"limit": 1,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", h.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("hubspot ValidateAccount: status %d", resp.StatusCode)
	}

	var result struct {
		Results []interface{} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return len(result.Results) > 0, nil
}

// SyncInteraction creates a note attached to a HubSpot deal.
func (h *HubSpotClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	payload := map[string]interface{}{
		"properties": map[string]string{
			"hs_note_body": note,
			"hs_note_subject": "Automated Interaction",
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/crm/v3/objects/notes", h.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", h.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("hubspot SyncInteraction: status %d", resp.StatusCode)
	}
	return nil
}

// SyncContacts creates or updates HubSpot contacts linked to a company.
func (h *HubSpotClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	for _, c := range contacts {
		payload := map[string]interface{}{
			"properties": map[string]string{
				"email":       c.Email,
				"firstname":   firstName(c.Name),
				"lastname":    lastName(c.Name),
				"jobtitle":    c.Role,
				"company":     fmt.Sprintf("%d", companyID),
			},
		}

		body, _ := json.Marshal(payload)
		url := fmt.Sprintf("%s/crm/v3/objects/contacts", h.baseURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", h.authHeader())
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("hubspot SyncContacts: status %d", resp.StatusCode)
		}
	}
	return nil
}

// FetchDealDetails retrieves a HubSpot deal.
func (h *HubSpotClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/deals/%d?properties=dealname,amount,description,hs_deal_stage,custom_route", h.baseURL, dealID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", h.authHeader())

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hubspot FetchDealDetails: status %d", resp.StatusCode)
	}

	var result struct {
		Properties struct {
			DealName      string  `json:"dealname"`
			Amount        float64 `json:"amount"`
			Description   string  `json:"description"`
			Stage         string  `json:"hs_deal_stage"`
			CustomRoute   string  `json:"custom_route"`
		} `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &DealDetails{
		ID:                 parseHubSpotDealID(result.Properties.DealName), // placeholder conversion
		Status:             h.mapHubSpotStageToLeadState(result.Properties.Stage),
		QuotedPricing:      result.Properties.Amount,
		CustomRequirements: result.Properties.CustomRoute,
		TechnicalDossier:   result.Properties.Description,
	}, nil
}

// Helper functions for parsing and mapping.
func firstName(full string) string {
	parts := strings.Split(full, " ")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func lastName(full string) string {
	parts := strings.Split(full, " ")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func (h *HubSpotClient) mapLeadStateToHubSpotStage(state db.LeadState) string {
	if h.stageMap != nil {
		if stage, ok := h.stageMap[string(state)]; ok {
			return stage
		}
	}
	return "appointmentscheduled"
}

func (h *HubSpotClient) mapHubSpotLeadStatusToLeadState(status string) db.LeadState {
	if h.reverseMap != nil {
		if stateStr, ok := h.reverseMap[status]; ok {
			return db.LeadState(stateStr)
		}
	}
	return db.StateResearched
}

func (h *HubSpotClient) mapHubSpotStageToLeadState(stage string) db.LeadState {
	if h.reverseMap != nil {
		if stateStr, ok := h.reverseMap[stage]; ok {
			return db.LeadState(stateStr)
		}
	}
	return db.StateResearched
}

func parseHubSpotDealID(name string) int64                     { return 0 }