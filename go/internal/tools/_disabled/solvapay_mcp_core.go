package tools

import (
	"context"
	"encoding/json"
)

// HandlePaywallMeta returns paywall metadata including CSP and bootstrap payload.
func HandlePaywallMeta(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	meta := map[string]interface{}{
		"domain": domain,
		"csp":    "default-src 'self'",
		"bootstrap": map[string]string{
			"url": "https://solvapay.com/paywall.js",
		},
	}
	data, e := json.Marshal(meta)
	if e != nil {
		return err(e.Error())
}

	return ok(string(data))
}

// HandleOAuthDiscovery returns OAuth discovery document.
func HandleOAuthDiscovery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issuer, _ :=getString(args, "issuer")
	if issuer == "" {
		return err("issuer is required")
}

	doc := map[string]interface{}{
		"issuer":                            issuer,
		"authorization_endpoint":            issuer + "/authorize",
		"token_endpoint":                    issuer + "/token",
		"jwks_uri":                          issuer + "/.well-known/jwks.json",
		"response_types_supported":          []string{"code", "token"},
		"subject_types_supported":           []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
	}
	data, e := json.Marshal(doc)
	if e != nil {
		return err(e.Error())
}

	return success(string(data))
}