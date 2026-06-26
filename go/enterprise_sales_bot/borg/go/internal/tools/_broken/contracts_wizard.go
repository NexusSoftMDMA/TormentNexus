package mcpimpl

import (
	"context"
	"strings"
	"time"
)

// HandleGenerateContract generates a simple contract from a template.
func HandleGenerateContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	partyName, _ :=getString(args, "party_name")
	contractType, _ :=getString(args, "contract_type")
	if partyName == "" || contractType == "" {
		return err("missing required parameters: party_name, contract_type")
}

	contract := "CONTRACT\nType: " + contractType + "\nParty: " + partyName + "\nDate: " + time.Now().Format(time.RFC3339)
	return ok("generated contract:\n" + contract)
}

// HandleAnalyzeContract analyzes a contract text for common terms.
func HandleAnalyzeContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("missing required parameter: text")
}

	analysis := []string{}
	if strings.Contains(strings.ToLower(text), "indemnification") {
		analysis = append(analysis, "indemnification clause found")

	if strings.Contains(strings.ToLower(text), "termination") {
		analysis = append(analysis, "termination clause found")

	if len(analysis) == 0 {
		analysis = append(analysis, "no standard clauses detected")

	return success("analysis:\n" + strings.Join(analysis, "\n"))
}
}
}
}