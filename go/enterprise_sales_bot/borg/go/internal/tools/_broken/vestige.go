package tools

import (
    "context"
    "fmt"
    "strings"
    "strconv"
    "time"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    limitStr, _ :=getString(args, "limit")
    limit := 10
    if limitStr != "" {
        if l, convErr := strconv.Atoi(limitStr); convErr == nil && l > 0 {
            limit = l
        }
    }
    // Simulate search result
    result := fmt.Sprintf("Search results for \"%s\" (limit %d):\n- Result 1\n- Result 2", query, limit)
    return ok(result)
}

func HandleAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    content, _ :=getString(args, "content")
    tagsStr, _ :=getString(args, "tags")
    tags := []string{}
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i := range tags {
            tags[i] = strings.TrimSpace(tags[i])

    }
    // Simulate adding memory
    msg := fmt.Sprintf("Added memory: \"%s\" with tags %v", content, tags)
    return ok(msg)
}

}

func HandlePurge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "id")
    if id == "" {
        return err("missing id")
    }
    // Simulate purge
    msg := fmt.Sprintf("Purged memory with id %s", id)
    return ok(msg)
}

func HandleListMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    limitStr, _ :=getString(args, "limit")
    limit := 5
    if limitStr != "" {
        if l, convErr := strconv.Atoi(limitStr); convErr == nil && l > 0 {
            limit = l
        }
    }
    // Simulate list
    msgs := []string{}
    for i := 1; i <= limit; i++ {
        msgs = append(msgs, fmt.Sprintf("Memory %d: example content", i))

    result := strings.Join(msgs, "\n")
    return ok(result)
}

}

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    info := "Vestige MCP server – local cognitive memory. Version 2.1.23."
    return ok(info)
}