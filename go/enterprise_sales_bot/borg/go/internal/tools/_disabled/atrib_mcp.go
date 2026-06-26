package tools

import "context"

func HandleSign(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    message, _ :=getString(args, "message")
    return ok("signed: " + message)
}

func HandleVerify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    signature, _ :=getString(args, "signature")
    return ok("verified: " + signature)
}