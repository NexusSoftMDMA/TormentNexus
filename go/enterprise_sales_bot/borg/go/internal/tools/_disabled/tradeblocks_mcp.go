package tools

import (
    "context"
    "strconv"
)

func HandleAnalyzeOption(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    symbol, _ :=getString(args, "symbol")
    strike, _ :=getInt(args, "strike")
    if symbol == "" {
        return err("symbol is required")
}

    return ok("Analyzed " + symbol + " at strike " + strconv.Itoa(strike))
}

func HandleGetOptionChain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    symbol, _ :=getString(args, "symbol")
    expiration, _ :=getString(args, "expiration")
    if symbol == "" {
        return err("symbol is required")
}

    if expiration == "" {
        return err("expiration is required")
}

    return ok("Option chain for " + symbol + " expiring " + expiration)
}