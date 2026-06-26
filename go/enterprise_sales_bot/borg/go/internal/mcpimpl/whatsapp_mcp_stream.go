package mcpimpl

import (
    "context"
)

func HandleSendMessage_whatsapp_mcp_stream(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    phone, _ :=getString(args, "phone")
    if phone == "" {
        return err("phone is required")
}

    message, _ :=getString(args, "message")
    if message == "" {
        return err("message is required")
}

    return ok("Message sent to " + phone + ": " + message)
}