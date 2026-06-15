package tools

import (
    "context"
    "fmt"
)

func HandleGetCaffeineInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        name = "Caffeine"
    }
    return ok(fmt.Sprintf("%s is a stimulant that boosts energy and alertness.", name))
}

func HandleRandomFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    facts := []string{
        "Caffeine can improve memory consolidation.",
        "Coffee, tea, and chocolate all contain caffeine.",
        "Caffeine blocks adenosine receptors to keep you awake.",
    }
    idx, _ :=getInt(args, "index")
    if idx < 0 || idx >= len(facts) {
        idx = 0
    }
    return ok(facts[idx])
}