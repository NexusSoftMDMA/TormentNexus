package tools

import (
    "context"
    "fmt"
    "regexp"
    "strings"
)

// ToolResponse, ok, e, getString, getInt, getBool, TextContent は既に定義されていると仮定

// HandleGenerateContext generates a context string.
func HandleGenerateContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    role, _ :=getString(args, "role")
    task, _ :=getString(args, "task")
    details, _ :=getString(args, "details")
    if role == "" || task == "" {
        return err("role and task are required")
    }
    contextStr := fmt.Sprintf("Role: %s\nTask: %s\nDetails: %s", role, task, details)
    return ok(contextStr)
}

// HandleFormatContext formats a context string.
func HandleFormatContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    contextStr, _ :=getString(args, "context")
    style, _ :=getString(args, "style")
    var formattedStr string
    switch style {
    case "upper":
        formattedStr = strings.ToUpper(contextStr)
    case "lower":
        formattedStr = strings.ToLower(contextStr)
    case "title":
        formattedStr = strings.Title(contextStr)
    default:
        formattedStr = contextStr // default no formatting
    }
    return ok(formattedStr)
}

// HandleExtractKeywords extracts keywords from text.
func HandleExtractKeywords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    text, _ :=getString(args, "text")
    if text == "" {
        return err("text is required")
    }
    re, compileErr := regexp.Compile(`\b\w+\b`)
    if compileErr != nil {
        return err("failed to compile regex: " + compileErr.Error())
    }
    matches := re.FindAllString(text, -1)
    keywords := strings.Join(matches, ", ")
    return ok(keywords)
}