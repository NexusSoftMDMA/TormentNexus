package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMarkets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	closed, _ :=getBool(args, "closed")
	url := fmt.Sprintf("https://clob.polymarket.com/markets?limit=%d&closed=%v", limit, closed)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleGetMarket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://clob.polymarket.com/markets/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}