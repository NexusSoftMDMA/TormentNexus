package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// http.DefaultClient returns a reusable HTTP client with a timeout.
func http.DefaultClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// HandleRegisterAgent registers a new agent.
// Expected args: "name" (string)
func HandleRegisterAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing required argument: name")
}

	// In a full implementation we would call the Prismer API here.
	// For now we simply acknowledge the request.
	return ok(fmt.Sprintf("Agent '%s' registered successfully.", name))
}

// HandleSendDirectMessage sends a direct message from one agent to another.
// Expected args: "from" (string), "to" (string), "message" (string)
func HandleSendDirectMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	message, _ :=getString(args, "message")

	if from == "" || to == "" || message == "" {
		return err("missing required arguments: from, to, and message must be provided")
}

	// Placeholder implementation – acknowledge the send request.
	return ok(fmt.Sprintf("Message from '%s' to '%s' queued: %s", from, to, message))
}

// HandleCreateGroup creates a group with given members.
// Expected args: "owner" (string), "title" (string), "description" (string, optional), "members" ([]interface{})
func HandleCreateGroup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	title, _ :=getString(args, "title")
	description, _ :=getString(args, "description")

	if owner == "" || title == "" {
		return err("missing required arguments: owner and title are required")
}

	// Members list is optional; we just stringify it for demonstration.
	membersRaw, found := args["members"]
	var membersStr string
	if found {
		switch v := membersRaw.(type) {
		case []interface{}:
			var list []string
			for _, m := range v {
				if s, found := m.(string); found {
					list = append(list, s)

			}
			membersStr = fmt.Sprintf("%v", list)
		default:
			membersStr = fmt.Sprintf("%v", v)

	}
	return ok(fmt.Sprintf("Group '%s' created by '%s'. Description: %s. Members: %s", title, owner, description, membersStr))
}

}
}

// HandlePresignUpload requests a presigned upload URL.
// Expected args: "fileName" (string), "fileSize" (int), "mimeType" (string)
func HandlePresignUpload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileName, _ :=getString(args, "fileName")
	mimeType, _ :=getString(args, "mimeType")
	fileSize, _ :=getInt(args, "fileSize")

	if fileName == "" || mimeType == "" || fileSize <= 0 {
		return err("missing or invalid arguments: fileName, mimeType, and positive fileSize are required")
}

	// Placeholder response mimicking a presign payload.
	uploadID := fmt.Sprintf("upload-%d", time.Now().Unix())
	url := fmt.Sprintf("https://example.prismer.cloud/upload/%s", uploadID)
	fields := map[string]string{
		"key":        uploadID,
		"Content-Type": mimeType,
	}
	// Encode fields as JSON string for the response.
	fieldsJSON, _ := json.Marshal(fields)

	return ok(fmt.Sprintf(`Presigned upload ready. uploadId: %s, url: %s, fields: %s`, uploadID, url, string(fieldsJSON)))
}

// HandleConfirmUpload confirms that an upload has completed.
// Expected args: "uploadId" (string)
func HandleConfirmUpload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uploadID, _ :=getString(args, "uploadId")
	if uploadID == "" {
		return err("missing required argument: uploadId")
}

	// Placeholder confirmation.
	cdnURL := fmt.Sprintf("https://cdn.prismer.cloud/files/%s", uploadID)
	return ok(fmt.Sprintf("Upload %s confirmed. CDN URL: %s", uploadID, cdnURL))
}