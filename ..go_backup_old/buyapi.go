package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleSearchVendors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query := url.QueryEscape(getString(args, "query"))
	if query == "" {
		return err("query parameter is required")
}

	resp, e := http.DefaultClient.Get("https://api.buyapi.dev/v1/search?q=" + query)
	if e != nil {
		return err("failed to call BuyAPI: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("BuyAPI returned status " + resp.Status)
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}