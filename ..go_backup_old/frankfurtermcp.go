package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGetLatestRates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base")
	symbols, _ :=getString(args, "symbols")
	u := "https://api.frankfurter.app/latest"
	params := url.Values{}
	if base != "" {
		params.Set("base", base)

	if symbols != "" {
		params.Set("symbols", symbols)

	if len(params) > 0 {
		u += "?" + params.Encode()

	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch rates: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("%v", data))
}

}
}
}

func HandleGetHistoricalRates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	base, _ :=getString(args, "base")
	symbols, _ :=getString(args, "symbols")
	if date == "" {
		return err("date is required")
}

	u := "https://api.frankfurter.app/" + date
	params := url.Values{}
	if base != "" {
		params.Set("base", base)

	if symbols != "" {
		params.Set("symbols", symbols)

	if len(params) > 0 {
		u += "?" + params.Encode()

	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch rates: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("%v", data))
}
}
}
}