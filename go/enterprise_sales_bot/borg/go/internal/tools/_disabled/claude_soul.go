package tools

import (
    "context"
)

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    qtype, _ :=getString(args, "type")
    if qtype == "" {
        return ok("The best way to predict the future is to create it. - Peter Drucker")
}

    if qtype == "famous" {
        return ok("I think, therefore I am. - Descartes")
}

    return ok("Claude Soul says: Stay curious.")
}

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return ok("Hello, Claude Soul!")
}

    return ok("Hello, " + name + "!")
}