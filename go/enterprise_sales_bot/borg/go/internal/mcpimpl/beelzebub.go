package mcpimpl

import (
    "context"
)

func HandleGreet_beelzebub(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    return success("Greetings from Beelzebub, " + name + "!")
}

func HandlePing_beelzebub(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}