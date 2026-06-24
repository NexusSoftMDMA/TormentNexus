package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func HandleConvert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	amountStr, _ :=getString(args, "amount")
	amount, e := strconv.ParseFloat(amountStr, 64)
	if e != nil {
		return err("invalid amount")
}

	rate, e := getExchangeRate(from, to)
	if e != nil {
		return err("rate fetch failed: " + e.Error())
}

	result := amount * rate
	return success(fmt.Sprintf("%.2f %s = %.2f %s", amount, from, result, to))
}

func HandleGetRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	rate, e := getExchangeRate(from, to)
	if e != nil {
		return err("rate fetch failed: " + e.Error())
}

	return success(fmt.Sprintf("1 %s = %.4f %s", from, rate, to))
}

func getExchangeRate(from, to string) (float64, error) {
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", from)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return 0, e
	}
	defer resp.Body.Close()
	var result struct {
		Rates map[string]float64 `json:"rates"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return 0, e
	}
	rate, found := result.Rates[to]
	if !found {
		return 0, fmt.Errorf("rate not found for %s", to)
}

	return rate, nil
}