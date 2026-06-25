package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func ok(content string) (ToolResponse, error) {
	return ToolResponse{Content: content}, nil
}

func err(message string) (ToolResponse, error) {
	return ToolResponse{Content: message}, fmt.Errorf(message)
}

func getString(args map[string]interface{}, key string) string {
	if val, found := args[key].(string); found {
		return val
	}
	return ""
}

func getInt(args map[string]interface{}, key string) int {
	if val, found := args[key].(float64); found {
		return int(val)
}

	return 0
}

func getBool(args map[string]interface{}, key string) bool {
	if val, found := args[key].(bool); found {
		return val
	}
	return false
}

func HandleInstallLinux(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	commands := []string{
		"sudo dnf install -y nss nspr mesa-libgbm libXcomposite libXdamage libXrandr alsa-lib pciutils-libs libXcursor atk at-spi2-atk pango cairo gdk-pixbuf2 gtk3",
		"sudo dnf install -y wl-clipboard",
		"sudo dnf install -y xclip",
	}
	for _, cmd := range commands {
		if e := exec.Command("bash", "-c", cmd).Run(); e != nil {
			return err(fmt.Sprintf("Failed to execute command: %s", cmd))

	}
	return ok("Linux installation completed successfully.")
}

}

func HandleInstallMac(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	commands := []string{
		"brew install node",
		"npm install -g @anthropic-ai/claude-code",
	}
	for _, cmd := range commands {
		if e := exec.Command("bash", "-c", cmd).Run(); e != nil {
			return err(fmt.Sprintf("Failed to execute command: %s", cmd))

	}
	return ok("macOS installation completed successfully.")
}

}

func HandleLaunchEditor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	editorPath, _ :=getString(args, "editorPath")
	if editorPath == "" {
		return err("Editor path is required.")
}

	cmd := exec.Command(editorPath, "-vulkan")
	if e := cmd.Start(); e != nil {
		return err(fmt.Sprintf("Failed to launch editor: %s", e.Error()))
}

	return ok("Editor launched successfully.")
}