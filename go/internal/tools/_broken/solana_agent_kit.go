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

func HandleCreateBugReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	description, _ :=getString(args, "description")
	steps, _ :=getString(args, "steps_to_reproduce")
	expected, _ :=getString(args, "expected_behavior")
	screenshots, _ :=getString(args, "screenshots")
	contextInfo, _ :=getString(args, "additional_context")

	var report strings.Builder
	report.WriteString("**Bug Report**\n\n")
	report.WriteString(fmt.Sprintf("**Title**: %s\n\n", title))
	report.WriteString("**Describe the bug**\n")
	report.WriteString(description + "\n\n")
	report.WriteString("**To Reproduce**\n")
	report.WriteString(steps + "\n\n")
	report.WriteString("**Expected behavior**\n")
	report.WriteString(expected + "\n\n")

	if screenshots != "" {
		report.WriteString("**Screenshots**\n")
		report.WriteString(screenshots + "\n\n")

	if contextInfo != "" {
		report.WriteString("**Additional context**\n")
		report.WriteString(contextInfo + "\n")

	return ok(report.String())
}

}
}

func HandleCreateFeatureRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	problem, _ :=getString(args, "problem_description")
	solution, _ :=getString(args, "solution_description")
	alternatives, _ :=getString(args, "alternatives_considered")
	contextInfo, _ :=getString(args, "additional_context")

	var request strings.Builder
	request.WriteString("**Feature Request**\n\n")
	request.WriteString(fmt.Sprintf("**Title**: %s\n\n", title))
	request.WriteString("**Is your feature request related to a problem?**\n")
	request.WriteString(problem + "\n\n")
	request.WriteString("**Describe the solution you'd like**\n")
	request.WriteString(solution + "\n\n")

	if alternatives != "" {
		request.WriteString("**Describe alternatives you've considered**\n")
		request.WriteString(alternatives + "\n\n")

	if contextInfo != "" {
		request.WriteString("**Additional context**\n")
		request.WriteString(contextInfo + "\n")

	return ok(request.String())
}

}
}

func HandleCreatePullRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issue, _ :=getString(args, "related_issue")
	changes, _ :=getString(args, "changes_made")
	details, _ :=getString(args, "implementation_details")
	transaction, _ :=getString(args, "example_transaction")
	prompt, _ :=getString(args, "prompt_used")
	notes, _ :=getString(args, "additional_notes")

	var pr strings.Builder
	pr.WriteString("# Pull Request Description\n\n")
	pr.WriteString(fmt.Sprintf("## Related Issue\nFixes #%s\n\n", issue))
	pr.WriteString("## Changes Made\nThis PR adds the following changes:\n")
	pr.WriteString(changes + "\n\n")
	pr.WriteString("## Implementation Details\n")
	pr.WriteString(details + "\n\n")

	if transaction != "" {
		pr.WriteString("## Transaction executed by agent\n")
		pr.WriteString("Example transaction:\n")
		pr.WriteString(transaction + "\n\n")

	if prompt != "" {
		pr.WriteString("## Prompt Used\n```\n")
		pr.WriteString(prompt + "\n```\n\n")

	if notes != "" {
		pr.WriteString("## Additional Notes\n")
		pr.WriteString(notes + "\n")

	pr.WriteString("## Checklist\n")
	pr.WriteString("- [ ] I have tested these changes locally\n")
	pr.WriteString("- [ ] I have updated the documentation\n")
	pr.WriteString("- [ ] I have added a transaction link\n")
	pr.WriteString("- [ ] I have added the prompt used to test it\n")

	return ok(pr.String())
}

}
}
}

func HandleRunBuildCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	packageName, _ :=getString(args, "package")
	command := "pnpm"
	var cmdArgs []string

	switch packageName {
	case "all":
		cmdArgs = []string{"run", "build"}
	case "core":
		cmdArgs = []string{"run", "build:core"}
	case "plugin-token":
		cmdArgs = []string{"run", "build:plugin-token"}
	case "plugin-defi":
		cmdArgs = []string{"run", "build:plugin-defi"}
	case "plugin-nft":
		cmdArgs = []string{"run", "build:plugin-nft"}
	case "plugin-misc":
		cmdArgs = []string{"run", "build:plugin-misc"}
	case "plugin-blinks":
		cmdArgs = []string{"run", "build:plugin-blinks"}
	case "adapter-mcp":
		cmdArgs = []string{"run", "build:adapter-mcp"}
	default:
		return err(fmt.Sprintf("unknown package: %s", packageName))
}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, command, cmdArgs...)
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("build failed: %v\n%s", cmdErr, string(output)))
}

	return ok(fmt.Sprintf("Build successful for %s:\n%s", packageName, string(output)))
}

func HandleRunLintCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fix, _ :=getBool(args, "fix")
	command := "pnpm"

	var cmdArgs []string
	if fix {
		cmdArgs = []string{"run", "lint:fix"}
	} else {
		cmdArgs = []string{"run", "lint"}
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, command, cmdArgs...)
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("lint failed: %v\n%s", cmdErr, string(output)))
}

	return ok(fmt.Sprintf("Lint %s:\n%s", map[bool]string{true: "fix", false: "check"}[fix], string(output)))
}

func HandleRunTestCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mode, _ :=getString(args, "mode")
	if mode != "agent" && mode != "programmatic" {
		return err("mode must be either 'agent' or 'programmatic'")
}

	command := "pnpm"
	cmdArgs := []string{"run", "test"}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, command, cmdArgs...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TEST_MODE=%s", mode),
	)

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("test failed: %v\n%s", cmdErr, string(output)))
}

	return ok(fmt.Sprintf("Test completed in %s mode:\n%s", mode, string(output)))
}