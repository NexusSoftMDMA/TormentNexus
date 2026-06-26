package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// BankModel represents a bank entity
type BankModel struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IsCompliant  bool   `json:"isCompliant"`
	BankTier     string `json:"bankTier"`
}

// PagingOfBankModel represents a paginated list of banks
type PagingOfBankModel struct {
	Count int          `json:"count"`
	Data  []BankModel  `json:"data"`
}

// Teller represents a bank teller
type Teller struct {
	GitHubProfile string `json:"gitHubProfile"`
}

// TellerReportList represents a list of reports
type TellerReportList struct {
	Count int    `json:"count"`
	Data  []struct {
		Name string `json:"name"`
	} `json:"data"`
}

// BankEvent represents a cloud event for bank operations
type BankEvent struct {
	SpecVersion    string      `json:"specversion"`
	ID             string      `json:"id"`
	Time           string      `json:"time"`
	Source         string      `json:"source"`
	Type           string      `json:"type"`
	DataContentType string     `json:"datacontenttype"`
	Data           interface{} `json:"data"`
}

// getBaseURL retrieves the base URL from environment or defaults
func getBaseURL() string {
	base := os.Getenv("BANK_API_BASE_URL")
	if base == "" {
		return "http://localhost:5201"
	}
	return base
}

// HandleListBanks retrieves a list of banks with optional pagination
func HandleListBanks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := getBaseURL()

	pageStr, _ :=getString(args, "page")
	limitStr, _ :=getString(args, "limit")

	page := 1
	limit := 21

	if pageStr != "" {
		if p, e := strconv.Atoi(pageStr); e == nil {
			page = p
		}
	}

	if limitStr != "" {
		if l, e := strconv.Atoi(limitStr); e == nil {
			limit = l
		}
	}

	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("limit", strconv.Itoa(limit))

	fullURL := fmt.Sprintf("%s/v1/banks?%s", baseURL, query.Encode())

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API Error %d: %s", resp.StatusCode, string(body)))
}

	var result PagingOfBankModel
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	jsonData, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(jsonData))
}

// HandleGetBank retrieves a specific bank by ID
func HandleGetBank(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bankID, _ :=getString(args, "id")
	if bankID == "" {
		return err("bank ID is required")
	}

	baseURL := getBaseURL()
	fullURL := fmt.Sprintf("%s/v1/banks/%s", baseURL, bankID)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API Error %d: %s", resp.StatusCode, string(body)))
	}

	var result BankModel
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
	}

	jsonData, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	return ok(string(jsonData))
}

// HandleCreateBank creates a new bank entry
func HandleCreateBank(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("bank name is required")
	}

	tier, _ :=getString(args, "tier")
	if tier == "" {
		tier = "B"
	}

	isCompliant, _ :=getBool(args, "isCompliant")

	bankData := map[string]interface{}{
		"name":         name,
		"bankTier":     tier,
		"isCompliant":  isCompliant,
	}

	jsonData, marshalErr := json.Marshal(bankData)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	baseURL := getBaseURL()
	fullURL := fmt.Sprintf("%s/v1/banks", baseURL)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(string(jsonData)))
	if reqErr != nil {
		return err(reqErr.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API Error %d: %s", resp.StatusCode, string(body)))
	}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
	}

	return ok(string(bodyBytes))
}

// HandleListTellers retrieves a list of bank tellers
func HandleListTellers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := getBaseURL()
	fullURL := fmt.Sprintf("%s/v1/tellers", baseURL)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API Error %d: %s", resp.StatusCode, string(body)))
	}

	var result []Teller
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
	}

	jsonData, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	return ok(string(jsonData))
}

// HandleGetTellerReport retrieves a specific teller report
func HandleGetTellerReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reportName, _ :=getString(args, "report_name")
	if reportName == "" {
		return err("report name is required")
	}

	baseURL := getBaseURL()
	fullURL := fmt.Sprintf("%s/v1/teller/reports/%s", baseURL, url.QueryEscape(reportName))

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API Error %d: %s", resp.StatusCode, string(body)))
	}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
	}

	return ok(string(bodyBytes))
}

// HandlePublishEvent publishes a bank event
func HandlePublishEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	eventType, _ :=getString(args, "type")
	bankID, _ :=getString(args, "bank_id")

	if eventType == "" {
		return err("event type is required")
	}
	if bankID == "" {
		return err("bank ID is required")
	}

	event := BankEvent{
		SpecVersion:     "1.0",
		ID:              fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Time:            time.Now().Format(time.RFC3339),
		Source:          "https://github.com/erwinkramer/bank-api",
		Type:            eventType,
		DataContentType: "application/json",
		Data: map[string]string{
			"bankId": bankID,
		},
	}

	jsonData, marshalErr := json.Marshal(event)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	baseURL := getBaseURL()
	fullURL := fmt.Sprintf("%s/v1/events", baseURL)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(string(jsonData)))
	if reqErr != nil {
		return err(reqErr.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API Error %d: %s", resp.StatusCode, string(body)))
	}

	return ok("Event published successfully")
}