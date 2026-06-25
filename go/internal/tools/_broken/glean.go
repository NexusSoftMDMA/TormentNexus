package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// HandleInfo returns a brief description of the Glean project.
func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	const description = "Glean is a self-hosted RSS reader and personal knowledge management tool."
	return ok(description)
}

// HandleDockerCompose fetches the requested docker-compose file from the Glean GitHub repository.
// Expected argument:
//   "variant" (string) – either "full" (default) or "lite". Determines which compose file to download.
func HandleDockerCompose(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	variant := strings.ToLower(getString(args, "variant"))
	if variant != "lite" {
		variant = "full"
	}

	var composePath string
	if variant == "lite" {
		composePath = "docker-compose.lite.yml"
	} else {
		composePath = "docker-compose.yml"
	}

	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/LeslieLeung/glean/main/%s", composePath)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to fetch compose file: HTTP %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleDeployCommand builds a ready-to-run docker-compose command based on supplied arguments
// and environment variables. Arguments (all optional):
//   "admin_user" (string) – admin username
//   "admin_pass" (string) – admin password
//   "secret_key" (string) – JWT secret key
//   "variant"    (string) – "full" (default) or "lite"
func HandleDeployCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	adminUser, _ :=getString(args, "admin_user")
	if adminUser == "" {
		adminUser = os.Getenv("ADMIN_USERNAME")

	adminPass, _ :=getString(args, "admin_pass")
	if adminPass == "" {
		adminPass = os.Getenv("ADMIN_PASSWORD")

	secretKey, _ :=getString(args, "secret_key")
	if secretKey == "" {
		secretKey = os.Getenv("SECRET_KEY")

	variant := strings.ToLower(getString(args, "variant"))
	if variant != "lite" {
		variant = "full"
	}

	_ = variant

	// Build a temporary .env content
	envLines := []string{
		fmt.Sprintf("ADMIN_USERNAME=%s", adminUser),
		fmt.Sprintf("ADMIN_PASSWORD=%s", adminPass),
		fmt.Sprintf("SECRET_KEY=%s", secretKey),
		"CREATE_ADMIN=true",
	}
	envContent := strings.Join(envLines, "\n")

	// Encode the .env content as a data URL so the user can pipe it directly.
	dataURL := "data:text/plain;base64," + base64.StdEncoding.EncodeToString([]byte(envContent))

	// Construct the final command.
	var cmdBuilder strings.Builder
	fmt.Fprintf(&cmdBuilder, "curl -fsSL %s -o docker-compose.yml && ", dataURL)
	fmt.Fprintf(&cmdBuilder, "curl -fsSL https://raw.githubusercontent.com/LeslieLeung/glean/main/.env.example -o .env && ")
	fmt.Fprintf(&cmdBuilder, "echo \"%s\" >> .env && ", escapeForShell(envContent))
	fmt.Fprintf(&cmdBuilder, "docker compose up -d")

	return ok(cmdBuilder.String())
}

}
}
}

// escapeForShell escapes newlines and double quotes for safe inclusion in a shell echo command.
func escapeForShell(s string) string {
	escaped := strings.ReplaceAll(s, `"`, `\"`)")
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	return escaped
}