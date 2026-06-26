package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:                "mcp",
	Short:              "MCP Router and Server management",
	Long:               `MCP Router and Server management. Proxies commands to the Go sidecar.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get current directory: %v\n", err)
			os.Exit(1)
		}

		// Dynamically find the sidecar path relative to the running executable directory first,
		// and fall back to the workspace bin/tormentnexus.exe.
		execPath, err := os.Executable()
		if err != nil {
			execPath = ""
		}
		
		var sidecarPath string
		if execPath != "" {
			sidecarPath = filepath.Join(filepath.Dir(execPath), "bin", "tormentnexus.exe")
			if _, err := os.Stat(sidecarPath); err == nil {
				goto found
			}
			sidecarPath = filepath.Join(filepath.Dir(execPath), "tormentnexus.exe")
			if _, err := os.Stat(sidecarPath); err == nil {
				goto found
			}
		}

		sidecarPath = filepath.Join(cwd, "bin", "tormentnexus.exe")
		if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
			sidecarPath = filepath.Join(cwd, "go", "bin", "tormentnexus.exe")
			if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Sidecar binary not found. Please run 'build.bat' or compile inside the 'go' folder first.\n")
				os.Exit(1)
			}
		}

	found:

		// Build proxy command
		subArgs := append([]string{"mcp"}, args...)
		c := exec.Command(sidecarPath, subArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		err = c.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if ws, ok := exitError.Sys().(syscall.WaitStatus); ok {
					os.Exit(ws.ExitStatus())
				}
				os.Exit(exitError.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "failed to run sidecar: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
