package mcpimpl

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleSubmitUrl_isiahw1_mcp_server_bing_webmaster(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	siteUrl, _ :=getString(args, "site_url")
	urlToSubmit, _ :=getString(args, "url")
	if apiKey == "" || siteUrl == "" || urlToSubmit == "" {
		return err("Missing required arguments: api_key, site_url, url")
}

	apiURL := "https://bingwebmaster.api/SubmitUrl?apikey=" + url.QueryEscape(apiKey) +
		"&siteUrl=" + url.QueryEscape(siteUrl) + "&url=" + url.QueryEscape(urlToSubmit)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("Read response error: " + e.Error())
}

	return ok(string(body))
}

func HandleGetSiteStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	siteUrl, _ :=getString(args, "site_url")
	if apiKey == "" || siteUrl == "" {
		return err("Missing required arguments: api_key, site_url")
}

	apiURL := "https://bingwebmaster.api/GetSiteStats?apikey=" + url.QueryEscape(apiKey) +
		"&siteUrl=" + url.QueryEscape(siteUrl)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("Read response error: " + e.Error())
}

	return ok(string(body))
}