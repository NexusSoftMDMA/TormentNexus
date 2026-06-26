package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func HandlePreRelease(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseTag, _ :=getString(args, "base_tag")
	updateReleaseNotes, _ :=getBool(args, "update_release_notes")
	updateContributors, _ :=getBool(args, "update_contributors")
	auditHelpDocs, _ :=getBool(args, "audit_help_docs")

	prevTag, apiErr := determinePreviousTag(baseTag)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var changedFiles []string
	if updateReleaseNotes {
		files, updateErr := updateReleaseNotesFile(prevTag)
		if updateErr != nil {
			return err(updateErr.Error())
}

		changedFiles = append(changedFiles, files...)

	if updateContributors {
		files, updateErr := updateContributorsFile(prevTag)
		if updateErr != nil {
			return err(updateErr.Error())
}

		changedFiles = append(changedFiles, files...)

	if auditHelpDocs {
		files, auditErr := auditHelpAndDocs(prevTag)
		if auditErr != nil {
			return err(auditErr.Error())
}

		changedFiles = append(changedFiles, files...)

	validation, validateErr := validateChanges(changedFiles)
	if validateErr != nil {
		return err(validateErr.Error())
}

	response := fmt.Sprintf(
		"Pre-release checks completed for range %s..HEAD\n"+
			"Files changed: %s\n"+
			"Validation: %s",
		prevTag,
		strings.Join(changedFiles, ", "),
		validation,
	)
	return ok(response)
}

}
}
}

func determinePreviousTag(baseTag string) (string, error) {
	if baseTag != "" {
		return baseTag, nil
	}
	cmd := exec.Command("git", "tag", "--list", "v*", "--sort=-version:refname")
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return "", fmt.Errorf("failed to list tags: %w", cmdErr)
}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
}

	return tags[0], nil
}

func updateReleaseNotesFile(prevTag string) ([]string, error) {
	changelogPath := "WHATSNEW.md"
	if _, statErr := os.Stat(changelogPath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("changelog file not found: %s", changelogPath)
}

	cmd := exec.Command("git", "log", "--oneline", fmt.Sprintf("%s..HEAD", prevTag))
	logOutput, cmdErr := cmd.Output()
	if cmdErr != nil {
		return nil, fmt.Errorf("failed to get git log: %w", cmdErr)
}

	content, readErr := os.ReadFile(changelogPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read changelog: %w", readErr)
}

	newContent := fmt.Sprintf("## %s\n%s\n\n%s", time.Now().Format("2006-01-02"), string(logOutput), string(content))
	if writeErr := os.WriteFile(changelogPath, []byte(newContent), 0644); writeErr != nil {
		return nil, fmt.Errorf("failed to write changelog: %w", writeErr)
}

	return []string{changelogPath}, nil
}

func updateContributorsFile(prevTag string) ([]string, error) {
	contributorsPath := "CONTRIBUTORS.md"
	if _, statErr := os.Stat(contributorsPath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("contributors file not found: %s", contributorsPath)
}

	cmd := exec.Command("git", "shortlog", "-sne", fmt.Sprintf("%s..HEAD", prevTag))
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return nil, fmt.Errorf("failed to get contributors: %w", cmdErr)
}

	content, readErr := os.ReadFile(contributorsPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read contributors file: %w", readErr)
}

	newContent := fmt.Sprintf("# Contributors\n\n%s\n\n%s", string(output), string(content))
	if writeErr := os.WriteFile(contributorsPath, []byte(newContent), 0644); writeErr != nil {
		return nil, fmt.Errorf("failed to write contributors file: %w", writeErr)
}

	return []string{contributorsPath}, nil
}

func auditHelpAndDocs(prevTag string) ([]string, error) {
	var changedFiles []string
	e := filepath.Walk("cmd", func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if strings.Contains(string(content), "Long:") {
				changedFiles = append(changedFiles, path)

		}
		return nil
	})
	if e != nil {
		return nil, fmt.Errorf("failed to audit help/docs: %w", e)
}

	return changedFiles, nil
}

}

func validateChanges(changedFiles []string) (string, error) {
	var validationResults []string
	for _, file := range changedFiles {
		if filepath.Ext(file) == ".go" {
			cmd := exec.Command("gofmt", "-w", file)
			if cmdErr := cmd.Run(); cmdErr != nil {
				return "", fmt.Errorf("gofmt failed on %s: %w", file, cmdErr)
}

			validationResults = append(validationResults, fmt.Sprintf("gofmt passed for %s", file))

	}

	cmd := exec.Command("git", "status", "--short")
	statusOutput, cmdErr := cmd.Output()
	if cmdErr != nil {
		return "", fmt.Errorf("failed to get git status: %w", cmdErr)
}

	validationResults = append(validationResults, fmt.Sprintf("git status:\n%s", string(statusOutput)))
	return strings.Join(validationResults, "\n"), nil
}

}

func HandleListTags(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("git", "tag", "--list", "v*", "--sort=-version:refname")
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return err(fmt.Sprintf("failed to list tags: %v", cmdErr))
}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 {
		return ok("No tags found")
}

	return ok("Tags:\n" + strings.Join(tags, "\n"))
}

func HandleGitLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base")
	head, _ :=getString(args, "head")
	if head == "" {
		head = "HEAD"
	}
	cmd := exec.Command("git", "log", "--oneline", fmt.Sprintf("%s..%s", base, head))
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return err(fmt.Sprintf("failed to get git log: %v", cmdErr))
}

	return ok("Commit log:\n" + string(output))
}

func HandleGitDiff(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base")
	head, _ :=getString(args, "head")
	if head == "" {
		head = "HEAD"
	}
	cmd := exec.Command("git", "diff", "--stat", base, head)
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return err(fmt.Sprintf("failed to get git diff: %v", cmdErr))
}

	return ok("Diff stat:\n" + string(output))
}

func HandleFindChangelogFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var changelogFiles []string
	e := filepath.Walk(".", func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() {
			lowerPath := strings.ToLower(path)
			if strings.Contains(lowerPath, "changelog") || strings.Contains(lowerPath, "whatsnew") {
				changelogFiles = append(changelogFiles, path)

		}
		return nil
	})
	if e != nil {
		return err(fmt.Sprintf("failed to find changelog files: %v", e))
}

	if len(changelogFiles) == 0 {
		return ok("No changelog files found")
}

	return ok("Changelog files:\n" + strings.Join(changelogFiles, "\n"))
}

}

func HandleFindHelpFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var helpFiles []string
	patterns := []string{"*.md", "*.go", "doc/*", "cmd/**/assets/*"}
	for _, pattern := range patterns {
		matches, globErr := filepath.Glob(pattern)
		if globErr != nil {
			continue
		}
		for _, match := range matches {
			if !strings.Contains(match, "test") && !strings.Contains(match, "_test") {
				helpFiles = append(helpFiles, match)

		}
	}

	sort.Strings(helpFiles)
	uniqueFiles := []string{}
	for i, file := range helpFiles {
		if i == 0 || file != helpFiles[i-1] {
			uniqueFiles = append(uniqueFiles, file)

	}

	if len(uniqueFiles) == 0 {
		return ok("No help files found")
}

	return ok("Help files:\n" + strings.Join(uniqueFiles, "\n"))
}
}
}