package mcpimpl

import "context"

func HandleGetCertification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account_id")
	certType, _ :=getString(args, "certification_type")
	if account == "" {
		return err("account_id is required")
}

	return ok("Account " + account + " has certification: " + certType)
}

func HandleIssueCertification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account_id")
	certType, _ :=getString(args, "certification_type")
	if account == "" {
		return err("account_id is required")
}

	if certType == "" {
		return err("certification_type is required")
}

	return success("Certification issued for " + account + " type " + certType)
}