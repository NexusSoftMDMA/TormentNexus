package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HandleGetIndicator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	indicator, _ :=getString(args, "indicator")
	country, _ :=getString(args, "country")
	start, _ :=getString(args, "startYear")
	end, _ :=getString(args, "endYear")
	if indicator == "" {
		return err("indicator parameter required")
}

	if country == "" {
		country = "all"
	}
	url := fmt.Sprintf("https://api.worldbank.org/v2/country/%s/indicator/%s?format=json&date=%s:%s&per_page=5000",
		country, indicator, start, end)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data []interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	if len(data) < 2 {
		return ok("no results found")
}

	entries, found := data[1].([]interface{})
	if !found {
		return ok("no entries returned")
}

	var values []string
	for _, entry := range entries {
		item, found := entry.(map[string]interface{})
		if !found {
			continue
		}
		if v, found := item["value"]; found && v != nil {
			year, _ := strconv.Atoi(strings.Split(item["date"].(string), "")[0])
			values = append(values, fmt.Sprintf("%d: %v", year, v))

	}
	if len(values) == 0 {
		return ok("no data values found")
}

	summary := fmt.Sprintf("Indicator: %s, Country: %s, Records: %d\n%s", indicator, country, len(values), strings.Join(values[:min(10, len(values))], "\n"))
	return ok(summary)
}

}

func HandleListCountries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.worldbank.org/v2/country?format=json&per_page=500"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data []interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	if len(data) < 2 {
		return ok("no countries")
}

	entries, found := data[1].([]interface{})
	if !found {
		return ok("no entries")
}

	var names []string
	for _, entry := range entries {
		item, found := entry.(map[string]interface{})
		if !found {
			continue
		}
		name, _ := item["name"].(string)
		code, _ := item["id"].(string)
		if name != "" && code != "" {
			names = append(names, code+" - "+name)

	}
	return ok(fmt.Sprintf("Countries (%d):\n%s", len(names), strings.Join(names[:min(20, len(names))], "\n")))
}

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}