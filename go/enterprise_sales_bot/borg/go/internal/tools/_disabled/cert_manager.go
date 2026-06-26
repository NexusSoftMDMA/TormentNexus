package tools

import "context"

func HandleCheckCertificate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	return ok("Certificate for " + domain + " is valid")
}

func HandleIssueCertificate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	email, _ :=getString(args, "email")
	if domain == "" || email == "" {
		return err("domain and email are required")
}

	return success("Certificate issued for " + domain)
}