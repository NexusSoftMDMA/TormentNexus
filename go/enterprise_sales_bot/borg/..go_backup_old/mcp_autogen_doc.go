package tools

import (
	"context"
	"fmt"
)

func HandleGenerateDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	doc := fmt.Sprintf("# Documentation for %s\n\nAuto-generated documentation for topic: %s", topic, topic)
	return success(doc)
}