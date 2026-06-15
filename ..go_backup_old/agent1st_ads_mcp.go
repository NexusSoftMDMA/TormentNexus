package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchAds(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://api.ads.example/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(fmt.Sprintf("search results: %s", string(body)))
}

func HandleGetAd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://api.ads.example/ad/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("get ad failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(fmt.Sprintf("ad details: %s", string(body)))
}