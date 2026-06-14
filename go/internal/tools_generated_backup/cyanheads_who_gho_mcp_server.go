package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleQueryGHO(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	country, _ :=getString(args, "country")
	region, _ :=getString(args, "region")
	year, _ :=getString(args, "year")
	sex, _ :=getString(args, "sex")

	url := fmt.Sprintf("https://ghoapi.azureedge.net/api/Indicator?$filter=GeoId eq '%s' and Region eq '%s' and Year eq '%s' and Sex eq '%s'",
		country, region, year, sex)
	// Use http.DefaultClient
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("WHO GHO data: %+v", result))
}