package mcp

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// DiscoveryServerLike represents a server config for preflight checks.
type DiscoveryServerLike struct {
	Name        string
	Type        string
	Command     string
	Args        []string
	Env         map[string]string
	URL         string
	Headers     map[string]string
	BearerToken string
}

// sampleValuePatterns are regex patterns that indicate placeholder/sample config values.
var sampleValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)YOUR_[A-Z0-9_]+_HERE`),
	regexp.MustCompile(`(?i)postgres://user:password@localhost:5432/dbname`),
	regexp.MustCompile(`(?i)Bearer\s+YOUR_[A-Z0-9_]+_HERE`),
}

// hasSampleValue checks if a config value contains placeholder/sample content.
func hasSampleValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for _, pattern := range sampleValuePatterns {
		if pattern.MatchString(trimmed) {
			return true
		}
	}
	return false
}

// findPlaceholderFields identifies which fields in a server config contain placeholders.
func findPlaceholderFields(server *DiscoveryServerLike) []string {
	var fields []string

	if hasSampleValue(server.Command) {
		fields = append(fields, "command")
	}

	for i, arg := range server.Args {
		if hasSampleValue(arg) {
			fields = append(fields, formatFieldIndex("args", i))
		}
	}

	for key, value := range server.Env {
		if hasSampleValue(value) {
			fields = append(fields, "env."+key)
		}
	}

	if hasSampleValue(server.URL) {
		fields = append(fields, "url")
	}

	for key, value := range server.Headers {
		if hasSampleValue(value) {
			fields = append(fields, "headers."+key)
		}
	}

	if hasSampleValue(server.BearerToken) {
		fields = append(fields, "bearerToken")
	}

	return fields
}

func formatFieldIndex(base string, index int) string {
	return fmt.Sprintf("%s[%d]", base, index)
}

// GetDiscoveryPreflightFailure checks whether a server is ready for discovery.
// Returns nil if the server passes all checks, or an error message describing the failure.
func GetDiscoveryPreflightFailure(server *DiscoveryServerLike) string {
	placeholderFields := findPlaceholderFields(server)
	if len(placeholderFields) > 0 {
		preview := strings.Join(placeholderFields[:min(4, len(placeholderFields))], ", ")
		if len(placeholderFields) > 4 {
			preview += ", ..."
		}
		return "Discovery skipped because " + server.Name +
			" still contains placeholder or sample configuration values (" +
			preview + "). Update the config and try again."
	}

	// For STDIO servers, check if command exists on PATH
	serverType := server.Type
	if serverType == "" {
		serverType = "STDIO"
	}

	if serverType == "STDIO" {
		command := strings.TrimSpace(server.Command)
		if command == "" {
			return ""
		}

		if !commandExists(command) {
			return "Discovery skipped because command \"" + command +
				"\" is not available on PATH for " + server.Name +
				". Install it or update the server command before retrying."
		}
	}

	return ""
}

// commandExists checks if a command is available on the system PATH.
func commandExists(command string) bool {
	// For Windows, we need to try with and without extension
	if runtime.GOOS == "windows" {
		_, err := exec.LookPath(command)
		if err == nil {
			return true
		}
		// Try common extensions
		for _, ext := range []string{".exe", ".bat", ".cmd", ".ps1"} {
			if _, err := exec.LookPath(command + ext); err == nil {
				return true
			}
		}
		return false
	}

	_, err := exec.LookPath(command)
	return err == nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
