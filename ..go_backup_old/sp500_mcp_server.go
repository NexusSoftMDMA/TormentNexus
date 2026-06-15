package tools

import (
	"context"
)

func HandleGetCompanyInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	data := `{"symbol":"` + symbol + `","name":"Example Corp","sector":"Technology"}`
	return ok(data)
}

func HandleGetCompanyNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	news := `[{"title":"Dummy News","date":"2023-01-01"}]`
	return success(news)
}