package mcpimpl

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleSpeak_speech_sh(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text required")
}

	cmd := exec.Command("say", text)
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("speak failed: %v", e))
}

	return ok("spoken")
}