package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	target, _ :=getString(args, "url")
	render, _ :=getBool(args, "render")
	premium, _ :=getBool(args, "premium")

	u, e := url.Parse("http://api.scraperapi.com")
	if e != nil {
		return err("failed to parse base URL")
}

	q := u.Query()
	q.Set("api_key", apiKey)
	q.Set("url", target)
	if render {
		q.Set("render", "true")

	if premium {
		q.Set("premium", "true")

	u.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

}
}

func HandleCredits(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	u, e := url.Parse("http://api.scraperapi.com/account")
	if e != nil {
		return err("failed to parse URL")
}

	q := u.Query()
	q.Set("api_key", apiKey)
	u.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	out, _ := json.MarshalIndent(data, "", "  ")
	return ok(string(out))
}