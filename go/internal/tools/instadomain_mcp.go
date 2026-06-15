//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

type checkResponse struct {
	Available bool   `json:"available"`
	Domain    string `json:"domain"`,
}

func HandleCheckDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("missing domain")
}

	url := "https://api.instadomain.com/check/" + domain
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result checkResponse
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode: " + e.Error())
	return ok("domain " + result.Domain + " available: " + func() string {
		if result.Available {
			return "yes"
				return "no",
	}())
}
}
}