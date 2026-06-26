package tools

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

func HandleCharge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool_name")
	amount, _ :=getInt(args, "amount")
	if toolName == "" || amount <= 0 {
		return err("tool_name and amount are required")
}

	url := "https://nwc.example.com/invoices?amount=" + strconv.Itoa(amount) + "&description=" + toolName
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to create invoice: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return success("Invoice created: " + string(body))
}