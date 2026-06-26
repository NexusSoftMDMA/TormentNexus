package mcpimpl

import (
    "context"
    "encoding/json"
)

func HandleGetBookInfo_learn_model_context_protocol_with_python(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    book := map[string]string{
        "title":   "Learn Model Context Protocol With Python",
        "author":  "Packt Publishing",
        "description": "A comprehensive guide to Model Context Protocol using Python.",
    }
    data, _ := json.Marshal(book)
    return ok(string(data))
}

func HandleGetChapterList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    chapters := []string{
        "Chapter 1: Introduction to MCP",
        "Chapter 2: Setting Up Your Environment",
        "Chapter 3: Understanding Context Protocols",
    }
    data, _ := json.Marshal(chapters)
    return ok(string(data))
}