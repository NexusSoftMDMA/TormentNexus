package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetEdgarFilings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	url := fmt.Sprintf("https://data.sec.gov/submissions/CIK000%010d.json", tickerToCIK(ticker))
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("User-Agent", "MCP-Server/1.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("EDGAR filings for %s: %s", ticker, string(body[:200])))
}

func tickerToCIK(ticker string) string {
	// Simplified mapping; in production use CIK lookup API
	m := map[string]string{
		"AAPL": "320193",
		"MSFT": "789019",
		"GOOG": "1652044",
	}
	if cik, found := m[ticker]; found {
		return cik
	}
	return "000000"
}