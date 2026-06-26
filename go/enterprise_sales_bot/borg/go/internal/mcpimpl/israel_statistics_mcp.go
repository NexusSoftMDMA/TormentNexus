package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleIsraelPopulation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	year, _ :=getString(args, "year")
	if year == "" {
		year = "2024"
	}
	resp, e := http.DefaultClient.Get("https://api.example.com/population?year=" + year)
	if e != nil {
		return err("failed to fetch population data")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	population, found := result["population"].(float64)
	if !found {
		return err("population not found in response")
}

	return ok("Israel population in " + year + ": " + formatFloat(population))
}

func formatFloat(f float64) string {
	out, _ := json.Marshal(f)
	return string(out)
}