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
	"sort"
	"strconv"
	"strings"
	"time"
)

func HandleListDags(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "directory")
	if dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".dagu", "dags")

	files, readErr := os.ReadDir(dir)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read dags directory: %v", readErr))
}

	var dagFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			dagFiles = append(dagFiles, strings.TrimSuffix(file.Name(), ".yaml"))

	}

	sort.Strings(dagFiles)
	return ok(strings.Join(dagFiles, "\n"))
}

}
}

func HandleGetDag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	dir := filepath.Join(os.Getenv("HOME"), ".dagu", "dags")
	filePath := filepath.Join(dir, name+".yaml")

	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read dag file: %v", readErr))
}

	return ok(string(content))
}

func HandleCreateDag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	content, _ :=getString(args, "content")
	if content == "" {
		content = `# Example DAG
steps:
  - name: step1
    command: echo "Hello World"
`
	}

	dir := filepath.Join(os.Getenv("HOME"), ".dagu", "dags")
	if e := os.MkdirAll(dir, 0755); e != nil {
		return err(fmt.Sprintf("failed to create dags directory: %v", e))
}

	filePath := filepath.Join(dir, name+".yaml")
	if e := os.WriteFile(filePath, []byte(content), 0644); e != nil {
		return err(fmt.Sprintf("failed to write dag file: %v", e))
}

	return ok(fmt.Sprintf("created dag %s", name))
}

func HandleDeleteDag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	dir := filepath.Join(os.Getenv("HOME"), ".dagu", "dags")
	filePath := filepath.Join(dir, name+".yaml")

	if e := os.Remove(filePath); e != nil {
		return err(fmt.Sprintf("failed to delete dag file: %v", e))
}

	return ok(fmt.Sprintf("deleted dag %s", name))
}

func HandleRunDag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	params, _ :=getString(args, "params")
	cmd := exec.CommandContext(ctx, "dagu", "start", name)
	if params != "" {
		cmd.Args = append(cmd.Args, "--params="+params)

	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("failed to run dag: %v\n%s", runErr, string(output)))
}

	return ok(string(output))
}

}

func HandleListRuns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	dir := filepath.Join(os.Getenv("HOME"), ".dagu", "history", name)
	files, readErr := os.ReadDir(dir)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return ok("no runs found")
}

		return err(fmt.Sprintf("failed to read runs directory: %v", readErr))
}

	var runs []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			runs = append(runs, strings.TrimSuffix(file.Name(), ".json"))

	}

	sort.Sort(sort.Reverse(sort.StringSlice(runs)))
	return ok(strings.Join(runs, "\n"))
}

}

func HandleGetRunStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	runID, _ :=getString(args, "run_id")
	if name == "" || runID == "" {
		return err("both name and run_id parameters are required")
}

	filePath := filepath.Join(os.Getenv("HOME"), ".dagu", "history", name, runID+".json")
	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read run status: %v", readErr))
}

	return ok(string(content))
}