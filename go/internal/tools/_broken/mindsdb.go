package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func HandleMindsDBPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL, _ :=getString(args, "api_url")
	if apiURL == "" {
		apiURL = "http://localhost:47334/api/ping"
	}

	client := http.DefaultClient
	resp, fetchErr := client.Get(apiURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to ping MindsDB: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("MindsDB ping failed with status: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read ping response: %v", readErr))
}

	return ok(string(body))
}

func HandleMindsDBQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL, _ :=getString(args, "api_url")
	if apiURL == "" {
		apiURL = "http://localhost:47334/api/query"
	}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	payload := map[string]interface{}{
		"query": query,
	}
	jsonData, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal query: %v", marshalErr))
}

	client := http.DefaultClient
	resp, postErr := client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	if postErr != nil {
		return err(fmt.Sprintf("failed to execute query: %v", postErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("query failed with status %s: %s", resp.Status, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read query response: %v", readErr))
}

	return ok(string(body))
}

func HandleMindsDBInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version == "" {
		version = "latest"
	}

	cmd := exec.Command("pip3", "install", fmt.Sprintf("mindsdb==%s", version))
	output, installErr := cmd.CombinedOutput()
	if installErr != nil {
		return err(fmt.Sprintf("failed to install MindsDB: %v\nOutput: %s", installErr, string(output)))
}

	return ok(fmt.Sprintf("MindsDB %s installed successfully", version))
}

func HandleMindsDBStartServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiPort, _ :=getString(args, "api_port")
	if apiPort == "" {
		apiPort = "47334"
	}

	studioPort, _ :=getString(args, "studio_port")
	if studioPort == "" {
		studioPort = "47335"
	}

	cmd := exec.Command("mindsdb", "--api=http", fmt.Sprintf("--port=%s", apiPort), fmt.Sprintf("--studio-port=%s", studioPort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startErr := cmd.Start()
	if startErr != nil {
		return err(fmt.Sprintf("failed to start MindsDB server: %v", startErr))
}

	go func() {
		_ = cmd.Wait()
	}()

	return ok(fmt.Sprintf("MindsDB server started on ports %s (API) and %s (Studio)", apiPort, studioPort))
}

func HandleMindsDBCheckInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("pip3", "show", "mindsdb")
	output, checkErr := cmd.CombinedOutput()
	if checkErr != nil {
		return err("MindsDB is not installed")
}

	versionRegex := regexp.MustCompile(`Version: (\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return ok("MindsDB is installed but version could not be determined")
}

	return ok(fmt.Sprintf("MindsDB is installed (version %s)", matches[1]))
}

func HandleMindsDBCloneRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoURL := "https://github.com/mindsdb/mindsdb.git"
	targetDir, _ :=getString(args, "target_dir")
	if targetDir == "" {
		targetDir = "mindsdb"
	}

	// Check if directory already exists
	if _, statErr := os.Stat(targetDir); statErr == nil {
		return err(fmt.Sprintf("directory %s already exists", targetDir))
}

	cmd := exec.Command("git", "clone", "--recurse-submodules", repoURL, targetDir)
	output, cloneErr := cmd.CombinedOutput()
	if cloneErr != nil {
		return err(fmt.Sprintf("failed to clone repository: %v\nOutput: %s", cloneErr, string(output)))
}

	return ok(fmt.Sprintf("MindsDB repository cloned to %s", targetDir))
}