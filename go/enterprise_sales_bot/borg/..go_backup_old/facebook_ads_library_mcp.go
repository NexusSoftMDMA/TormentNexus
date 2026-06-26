package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type fbAd struct {
	ID             string `json:"id"`
	AdCreativeBody string `json:"ad_creative_body"`
}

type fbResponse struct {
	Data []fbAd `json:"data"`
}

func HandleSearchAds(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("access_token is required")
}

	country, _ :=getString(args, "country")
	if country == "" {
		country = "US"
	}
	vals := url.Values{}
	vals.Set("search_terms", keyword)
	vals.Set("ad_type", "ALL")
	vals.Set("ad_reached_countries", "["+country+"]")
	vals.Set("access_token", token)
	u := "https://graph.facebook.com/v19.0/ads_archive?" + vals.Encode()
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var fbResp fbResponse
	e = json.Unmarshal(body, &fbResp)
	if e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	out, e := json.MarshalIndent(fbResp.Data, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	return ok(string(out))
}