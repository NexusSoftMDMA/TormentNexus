package tools

import (
    "context"
    "time"
)

// HandleCelestineGreet greets a user with a personalized message.
// It takes an optional "name" argument; defaults to "World".
func HandleCelestineGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "World"
    }
    return ok("Hello, " + name + "!")
}

// HandleCelestineEcho returns the provided message back to the user.
// Requires a "message" argument.
func HandleCelestineEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    msg, _ :=getString(args, "message")
    if msg == "" {
        return err("message is required")
}

    return ok(msg)
}

// HandleCelestineTime returns the current server time in RFC1123 format.
func HandleCelestineTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    now := time.Now().Format(time.RFC1123)
    return ok("Current time: " + now)
}