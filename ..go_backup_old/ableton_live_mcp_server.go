package tools

import (
	"context"
	"fmt"
)

func HandlePlay(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	track, _ :=getString(args, "track")
	if track == "" {
		return err("track argument required")
}

	return ok(fmt.Sprintf("Playing track %s", track))
}

func HandleStop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Playback stopped")
}