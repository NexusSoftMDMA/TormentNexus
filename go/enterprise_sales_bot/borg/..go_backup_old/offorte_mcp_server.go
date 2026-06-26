package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetOffers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	url := "https://api.offorte.com/offers?category=" + category
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch offers: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}