package tools

import (
	"fmt"
	"os/exec"
	"runtime"
)

func (r *Registry) registerGUITools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "launch_webview",
		Description: "Launches a system webview or browser window (Electron-Orchestrator parity). Arguments: url (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			url, _ := args["url"].(string)

			var cmd string
			var osArgs []string

			switch runtime.GOOS {
			case "windows":
				cmd = "cmd"
				osArgs = []string{"/c", "start", url}
			case "darwin":
				cmd = "open"
				osArgs = []string{url}
			default: // linux
				cmd = "xdg-open"
				osArgs = []string{url}
			}

			if err := exec.Command(cmd, osArgs...).Start(); err != nil {
				return "", fmt.Errorf("failed to launch window: %v", err)
			}

			return fmt.Sprintf("Launched window for %s", url), nil
		},
	})
}
