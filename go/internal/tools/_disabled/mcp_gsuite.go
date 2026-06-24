package tools

import (
    "context"
)

func HandleGmailList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query != "" {
        return ok("listing emails with query: " + query)
}

    return ok("listing all emails")
}

func HandleCalendarList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    maxResults, _ :=getInt(args, "maxResults")
    if maxResults > 0 {
        return ok("listing up to " + string(rune('0'+maxResults%10)) + " calendar events")
}

    return ok("listing calendar events")
}