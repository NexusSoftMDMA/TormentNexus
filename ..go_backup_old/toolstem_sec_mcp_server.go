package tools

import (
	"context"
	"net/http"
)

func HandleInsiderSignals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker required")
}

	resp, e := http.DefaultClient.Get("https://fake.sec.gov/insider/" + ticker)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("Insider signals for " + ticker)
}

func HandleInstitutionalFlow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker required")
}

	resp, e := http.DefaultClient.Get("https://fake.sec.gov/institutional/" + ticker)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("Institutional flow for " + ticker)
}