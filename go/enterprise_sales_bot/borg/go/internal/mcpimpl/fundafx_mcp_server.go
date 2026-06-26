package mcpimpl

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleCentralBankRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	country, _ :=getString(args, "country")
	url := "https://api.exchangerate-api.com/v4/latest/" + country
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch rates: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("Central bank rate data: " + string(body))
}

func HandleEconomicIndicator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	indicator, _ :=getString(args, "indicator")
	country, _ :=getString(args, "country")
	url := "https://api.exchangerate-api.com/v4/latest/" + country + "?indicator=" + indicator
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch indicator: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success("Economic indicator data: " + string(body))
}