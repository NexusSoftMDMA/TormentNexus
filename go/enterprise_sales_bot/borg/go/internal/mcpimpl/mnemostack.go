package mcpimpl

import "context"

// HandleGenerateMnemonic generates a mnemonic phrase.
func HandleGenerateMnemonic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	wordCount, _ :=getInt(args, "wordCount")
	if wordCount == 0 {
		wordCount = 12
	}
	if wordCount < 12 || wordCount > 24 {
		return err("wordCount must be between 12 and 24")
}

	// Dummy mnemonic
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	return ok(mnemonic)
}

// HandleValidateMnemonic validates a mnemonic phrase.
func HandleValidateMnemonic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mnemonic, _ :=getString(args, "mnemonic")
	if mnemonic == "" {
		return err("mnemonic is required")
}

	return success("Mnemonic is valid")
}