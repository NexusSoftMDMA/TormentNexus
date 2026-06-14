package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetSalesReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	storeID, _ :=getString(args, "store_id")
	url := fmt.Sprintf("https://api.example.com/sales?date=%s&store=%s", date, storeID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch sales report: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleGetAnalytics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	period, _ :=getString(args, "period")
	url := fmt.Sprintf("https://api.example.com/analytics?period=%s", period)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch analytics: " + e.Error())
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

	return ok("analytics retrieved")
}