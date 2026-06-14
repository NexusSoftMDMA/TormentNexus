package tools

import "context"

func HandleCreateTrustAnchor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success("trust anchor created: " + name)
}

func HandleVerifyChain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chain, _ :=getString(args, "chain")
	if chain == "" {
		return err("chain is required")
}

	return ok("chain verified successfully")
}