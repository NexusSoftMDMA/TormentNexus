package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleLookupNhi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nhi, _ :=getString(args, "nhi_number")
	if nhi == "" {
		return err("nhi_number is required")
}

	url := "https://api.opdstar.com/nhi/" + nhi
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}

func HandleValidateNhi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nhi, _ :=getString(args, "nhi_number")
	if len(nhi) != 7 {
		return err("NHI number must be 7 characters")
}

	// simple check: first char letter, rest digits
	found := len(nhi) > 0 && nhi[0] >= 'A' && nhi[0] <= 'Z'
	if !found {
		return err("first character must be uppercase letter")
}

	for i := 1; i < 7; i++ {
		if nhi[i] < '0' || nhi[i] > '9' {
			return err("characters 2-7 must be digits")

	}
	return success("valid NHI number format")
}
}