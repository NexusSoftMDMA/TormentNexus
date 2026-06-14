package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateOption(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	underlying, _ :=getString(args, "underlying")
	strike, _ :=getString(args, "strike")
	expiry, _ :=getString(args, "expiry")
	optionType, _ :=getString(args, "type")
	amount, _ :=getString(args, "amount")

	url := fmt.Sprintf("https://api.skew.fi/v1/options/create?underlying=%s&strike=%s&expiry=%s&type=%s&amount=%s",
		underlying, strike, expiry, optionType, amount)

	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request")
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}

	return success(fmt.Sprintf("Option created: %v", result))
}

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	underlying, _ :=getString(args, "underlying")
	strike, _ :=getString(args, "strike")
	expiry, _ :=getString(args, "expiry")
	optionType, _ :=getString(args, "type")

	url := fmt.Sprintf("https://api.skew.fi/v1/options/quote?underlying=%s&strike=%s&expiry=%s&type=%s",
		underlying, strike, expiry, optionType)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}

	return ok(fmt.Sprintf("Quote: %v", result))
}

func HandleSettleOption(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	optionID, _ :=getString(args, "option_id")

	url := fmt.Sprintf("https://api.skew.fi/v1/options/settle/%s", optionID)

	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request")
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}

	return success(fmt.Sprintf("Option settled: %v", result))
}// touch 1781132141
