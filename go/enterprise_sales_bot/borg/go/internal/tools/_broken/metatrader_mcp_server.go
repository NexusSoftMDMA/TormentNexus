package mcpimpl

import "net/http"

func HandleGetAccountInfo_metatrader_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url") + "/account"
	if url == "" {
		return err("url argument is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get account info: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("MetaTrader API returned status " + resp.Status)
}

	return success("Account info retrieved successfully")
}

func HandlePlaceOrder_metatrader_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	symbol, _ :=getString(args, "symbol")
	volume, _ :=getInt(args, "volume")
	if url == "" || symbol == "" || volume <= 0 {
		return err("url, symbol, and volume > 0 are required")
}

	reqURL := url + "/order?" + "symbol=" + symbol + "&volume=" + string(volume)
	resp, e := http.DefaultClient.Post(reqURL, "application/json", nil)
	if e != nil {
		return err("failed to place order: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("MetaTrader API returned status " + resp.Status)
}

	return ok("Order placed for " + symbol + " with volume " + string(volume))
}