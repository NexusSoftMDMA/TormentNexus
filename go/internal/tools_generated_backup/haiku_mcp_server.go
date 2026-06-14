package tools

import (
	"context"
	"fmt"
)

func HandleGenerateHaiku(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return success("An old silent pond\nA frog jumps into the pond\nSplash! Silence again.")
}

	haiku := fmt.Sprintf("On %s\nSomething happens here\nBut I'm not creative", topic)
	return success(haiku)
}