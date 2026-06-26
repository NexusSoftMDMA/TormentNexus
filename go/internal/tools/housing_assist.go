package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ok(message string) (ToolResponse, error) {
	return ToolResponse{Message: message}, nil
}

func err(message string) (ToolResponse, error) {
	return ToolResponse{Message: message}, fmt.Errorf(message)
}

func getString(args map[string]interface{}, key string) string {
	if value, found := args[key].(string); found {
		return value
	}
	return ""
}

func getInt(args map[string]interface{}, key string) int {
	if value, found := args[key].(float64); found {
		return int(value)
}

	return 0
}

func getBool(args map[string]interface{}, key string) bool {
	if value, found := args[key].(bool); found {
		return value
	}
	return false
}

// HandleFindHousing searches for housing listings based on location and budget.
// Expected args:
//   - location (string): city or area to search.
//   - budget (int): maximum monthly rent.
// Returns a formatted list of available listings.
func HandleFindHousing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	budget, _ :=getInt(args, "budget")

	// Prepare request
	baseURL := "https://api.example.com/housing"
	values := url.Values{}
	values.Set("location", location)
	values.Set("budget", strconv.Itoa(budget))
	fullURL := fmt.Sprintf("%s?%s", baseURL, values.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("housing service returned status %d", resp.StatusCode))
}

	type listing struct {
		ID      string `json:"id"`
		Address string `json:"address"`
		Price   int    `json:"price"`
	}
	var listings []listing
	decodeErr := json.NewDecoder(resp.Body).Decode(&listings)
	if decodeErr != nil {
		return err(decodeErr.Error())
}

	if len(listings) == 0 {
		return ok("No housing listings found for the given criteria.")
}

	var sb strings.Builder
	sb.WriteString("Available housing listings:\n")
	for _, l := range listings {
		sb.WriteString(fmt.Sprintf("- ID: %s | Address: %s | Price: $%d\n", l.ID, l.Address, l.Price))

	return ok(sb.String())
}

}

// HandleApplyHousing submits an application for a specific housing listing.
// Expected args:
//   - applicant_name (string): name of the applicant.
//   - applicant_income (int): monthly income of the applicant.
//   - listing_id (string): ID of the housing listing to apply for.
// Returns a confirmation message.
func HandleApplyHousing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "applicant_name")
	income, _ :=getInt(args, "applicant_income")
	listingID, _ :=getString(args, "listing_id")

	// Simple eligibility check before sending request
	if income < 2000 {
		return err("Applicant income is below the minimum required for housing assistance.")
}

	// Prepare request payload
	payload := map[string]string{
		"name":       name,
		"income":     strconv.Itoa(income),
		"listing_id": listingID,
	}
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.example.com/housing/apply", strings.NewReader(string(payloadBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("application failed with status %d", resp.StatusCode))
}

	type applyResp struct {
		ApplicationID string `json:"application_id"`
		Status        string `json:"status"`
	}
	var ar applyResp
	decodeErr := json.NewDecoder(resp.Body).Decode(&ar)
	if decodeErr != nil {
		return err(decodeErr.Error())
}

	msg := fmt.Sprintf("Application submitted successfully. Application ID: %s, Status: %s", ar.ApplicationID, ar.Status)
	return ok(msg)
}

// HandleCheckEligibility determines if a user qualifies for housing assistance.
// Expected args:
//   - household_income (int): total household monthly income.
//   - household_size (int): number of people in the household.
// Returns a simple eligibility statement.
func HandleCheckEligibility(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	income, _ :=getInt(args, "household_income")
	size, _ :=getInt(args, "household_size")

	// Basic eligibility rule: income per person must be below $1500
	if size <= 0 {
		return err("household_size must be greater than zero")
}

	threshold := 1500 * size
	if income <= threshold {
		return ok("The household is eligible for housing assistance.")
}

	return ok("The household does not meet the eligibility criteria for housing assistance.")
}