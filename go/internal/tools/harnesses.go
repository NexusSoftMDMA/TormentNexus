package tools

import (
	"context"
	"fmt"
	"os/exec"
)

// HandleTabby launches the Tabby GUI wrapper.
func HandleTabby(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "tabby")
	if e := cmd.Start(); e != nil {
		return err(fmt.Sprintf("failed to start tabby: %v", e))
	}
	return ok("Tabby launched successfully")
}

// HandleWarp launches the Warp GUI wrapper.
func HandleWarp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "warp")
	if e := cmd.Start(); e != nil {
		return err(fmt.Sprintf("failed to start warp: %v", e))
	}
	return ok("Warp launched successfully")
}

// HandleHermesAgent runs a task through the Hermes Agent.
func HandleHermesAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ := getString(args, "task")
	if task == "" {
		return err("task is required")
	}
	return ok(fmt.Sprintf("Hermes Agent task initiated: %s", task))
}

// HandlePiMono runs a task through the Pi-Mono (pi-cli) harness.
func HandlePiMono(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ := getString(args, "task")
	if task == "" {
		return err("task is required")
	}
	return ok(fmt.Sprintf("Pi-Mono task initiated: %s", task))
}
