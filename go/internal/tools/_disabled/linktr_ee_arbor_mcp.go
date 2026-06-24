package tools

import (
	"context"
	"fmt"
)

// HandleOpenInPlayroom opens a component in Arbor's Playroom.
func HandleOpenInPlayroom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	component, _ :=getString(args, "component")
	if component == "" {
		return err("component is required")
	}
	url := fmt.Sprintf("https://playroom.arbor.design/#?code=%s", component)
	msg := fmt.Sprintf("Open %s in Playroom: %s", component, url)
	return success(msg)
}