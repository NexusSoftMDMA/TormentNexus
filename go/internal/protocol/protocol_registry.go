package protocol

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// RegisterProtocol registers the tormentnexus:// protocol handler in the OS.
func RegisterProtocol() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("protocol registration is only supported on Windows")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath = filepath.Clean(execPath)

	// Register under HKCU\Software\Classes\tormentnexus so it doesn't require Admin privileges
	commands := [][]string{
		{"add", `HKCU\Software\Classes\tormentnexus`, "/ve", "/t", "REG_SZ", "/d", "URL:TormentNexus Protocol", "/f"},
		{"add", `HKCU\Software\Classes\tormentnexus`, "/v", "URL Protocol", "/t", "REG_SZ", "/d", "", "/f"},
		{"add", `HKCU\Software\Classes\tormentnexus\shell\open\command`, "/ve", "/t", "REG_SZ", "/d", fmt.Sprintf(`"%s" "%%1"`, execPath), "/f"},
	}

	for _, args := range commands {
		cmd := exec.Command("reg", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("reg command failed (args: %v, output: %s): %w", args, string(output), err)
		}
	}

	return nil
}
