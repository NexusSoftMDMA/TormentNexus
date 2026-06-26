package tools

import (
    "context"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    return success("Greetings from Beelzebub, " + name + "!")
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}