package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListLicenses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	if user == "" {
		return err("missing user parameter")
}

	url := fmt.Sprintf("https://go.netlicensing.io/api/v2/licensee/%s/license", user)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(getString(args, "apiKey"), "")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("Licenses: " + string(body))
}

func HandleValidateLicense(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	licenseNumber, _ :=getString(args, "licenseNumber")
	productNumber, _ :=getString(args, "productNumber")
	if licenseNumber == "" || productNumber == "" {
		return err("missing licenseNumber or productNumber")
}

	url := fmt.Sprintf("https://go.netlicensing.io/api/v2/licensee/validate?licenseNumber=%s&productNumber=%s", licenseNumber, productNumber)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(getString(args, "apiKey"), "")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	return ok("Validation result: " + string(body))
}