package tools', no markdown, no explanation. The developer instruction says "Output ONLY fixed Go code starting with 'package tools'. No explanation. No markdown."

Thus we should output the Go file content only, not the manifest. The user says "Fix ALL compilation errors in this Go code." The developer instruction says output only fixed Go code. So we should output the Go code only, not the manifest. The user didn't ask for manifest. So we follow developer instruction.

Thus output:

package tools

import (
    "context"
    "fmt"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    to, _ :=getString(args, "to")
    subject, _ :=getString(args, "subject")
    body, _ :=getString(args, "body")
    _ = body
    return ok(fmt.Sprintf("Email sent to %s with subject %s", to, subject))
}

func HandleListEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    folder, _ :=getString(args, "folder")
    return ok(fmt.Sprintf("Listing emails from folder %s", folder))
}

func HandleCreateEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    title, _ :=getString(args, "title")
    start, _ :=getString(args, "start")
    end, _ :=getString(args, "end")
    return ok(fmt.Sprintf("Event created: %s from %s to %s", title, start, end))
}

func HandleListEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    calendar, _ :=getString(args, "calendar")
    return ok(fmt.Sprintf("Listing events from calendar %s", calendar))
}

Make sure no extra spaces or markdown.package tools

import (
	"context"
	"fmt"
)
