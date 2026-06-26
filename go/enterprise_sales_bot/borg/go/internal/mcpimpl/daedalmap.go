package mcpimpl

import (
	"context"
	"net/http"
)

func HandleDaedalmap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient // ensure import
	queryType, _ :=getString(args, "type")
	var msg string
	switch queryType {
	case "volcanoes":
		msg = "Volcanoes data: free query. Use 'type=volcanoes'."
	case "currency":
		msg = "Currency time series: free query. Use 'type=currency'."
	case "earthquakes":
		msg = "Earthquakes data requires payment via x402 on Base USDC."
	case "tsunamis":
		msg = "Tsunamis data requires payment via x402 on Base USDC."
	default:
		msg = "Daedalmap: Free discovery of geographic data. Paid earthquakes/tsunamis via x402 (Base USDC)."
	}
	return ok(msg)
}