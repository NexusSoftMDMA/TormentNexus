package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	versionPattern = regexp.MustCompile(`\d+\.\d+\.\d+`)
)

// HandleGetVersionInfo returns current Solon version information
func HandleGetVersionInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version == "" {
		version = "latest"
	}

	return ok(fmt.Sprintf("Solon framework version: %s\nSupported JDK versions: 8, 11, 17, 21, 25\nRepository: https://github.com/opensolon/solon", version))
}

// HandleGenerateIssueTemplate creates a new issue template based on issue type
func HandleGenerateIssueTemplate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueType, _ :=getString(args, "type")
	description, _ :=getString(args, "description")
	version, _ :=getString(args, "version")

	if version == "" {
		version = "unknown"
	}

	var template strings.Builder

	switch strings.ToLower(issueType) {
	case "bug":
		template.WriteString("### 问题描述\n")
		template.WriteString(description + "\n\n")
		template.WriteString("### 我当前使用 Solon 版本是?\n")
		template.WriteString(version + "\n\n")
		template.WriteString("### 如何复现\n1.\n2.\n3.\n```java\n// 可在此输入示例代码\n```\n\n")
		template.WriteString("### 预期结果\n\n")
		template.WriteString("### 实际结果\n\n")
	case "feature":
		template.WriteString("### 关联版本\n")
		template.WriteString(version + "\n\n")
		template.WriteString("### 请描述您的需求或者改进建议\n")
		template.WriteString(description + "\n\n")
		template.WriteString("### 请描述你建议的实现方案\n\n")
	case "question":
		template.WriteString("### 关联版本\n")
		template.WriteString(version + "\n\n")
		template.WriteString("### 请描述您的问题\n")
		template.WriteString(description + "\n")
	default:
		return err("invalid issue type. Must be 'bug', 'feature', or 'question'")
}

	return ok(template.String())
}

// HandleValidateContribution checks if contribution follows guidelines
func HandleValidateContribution(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	changeDescription, _ :=getString(args, "change_description")
	commitMessage, _ :=getString(args, "commit_message")
	hasTests, _ :=getBool(args, "has_tests")
	followsConvention, _ :=getBool(args, "follows_convention")
	considersDocs, _ :=getBool(args, "considers_docs")

	var issues []string

	if !versionPattern.MatchString(changeDescription) {
		issues = append(issues, "change description should mention Solon version")

	if !strings.HasPrefix(commitMessage, "[") || !strings.Contains(commitMessage, "]") {
		issues = append(issues, "commit message should follow conventional commits format")

	if !hasTests {
		issues = append(issues, "missing test coverage")

	if !followsConvention {
		issues = append(issues, "commit message doesn't follow conventional commits")

	if !considersDocs {
		issues = append(issues, "documentation impact not considered")

	if len(issues) > 0 {
		return ok("Contribution validation issues:\n- " + strings.Join(issues, "\n- "))
}

	return ok("Contribution follows all guidelines ✓")
}

}
}
}
}
}

// HandleListRepositories returns list of Solon related repositories
func HandleListRepositories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoType, _ :=getString(args, "type")
	var repos []string

	switch strings.ToLower(repoType) {
	case "main":
		repos = []string{
			"solon - Main code repository",
			"solon-examples - Official website supporting sample code repository",
		}
	case "ai":
		repos = []string{
			"solon-ai - Solon Ai code repository",
		}
	case "cloud":
		repos = []string{
			"solon-cloud - Solon Cloud code repository",
			"solon-admin - Solon Admin code repository",
		}
	case "plugins":
		repos = []string{
			"solon-maven-plugin - Solon Maven plugin",
			"solon-gradle-plugin - Solon Gradle plugin",
			"solon-idea-plugin - Solon Idea plugin",
			"solon-vscode-plugin - Solon VsCode plugin",
		}
	case "jakarta":
		repos = []string{
			"solon-java17 - Solon Jakarta (base java17)",
			"solon-java25 - Solon Jakarta (base java25)",
		}
	case "all":
		repos = []string{
			"solon - Main code repository",
			"solon-examples - Official website supporting sample code repository",
			"solon-ai - Solon Ai code repository",
			"solon-flow - Solon Flow code repository",
			"solon-expression - Solon Expression code repository",
			"solon-cloud - Solon Cloud code repository",
			"solon-admin - Solon Admin code repository",
			"solon-integration - Solon Integration code repository",
			"solon-java17 - Solon Jakarta (base java17)",
			"solon-java25 - Solon Jakarta (base java25)",
			"soloncode - SolonCode (Java8 impl version of Claude Code)",
			"solonclaw - SolonClaw (Java8 impl version of OpenClaw)",
			"solon-maven-plugin - Solon Maven plugin",
			"solon-gradle-plugin - Solon Gradle plugin",
			"solon-idea-plugin - Solon Idea plugin",
			"solon-vscode-plugin - Solon VsCode plugin",
			"solon-plugins - Third-party extension plugins",
		}
	default:
		return err("invalid repository type. Must be 'main', 'ai', 'cloud', 'plugins', 'jakarta', or 'all'")
}

	return ok(strings.Join(repos, "\n"))
}

// HandleGeneratePRTemplate creates a pull request template
func HandleGeneratePRTemplate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	purpose, _ :=getString(args, "purpose")
	summary, _ :=getString(args, "summary")

	if purpose == "" {
		return err("purpose is required")
}

	if summary == "" {
		return err("summary is required")
}

	template := fmt.Sprintf(`### 这个PR有什么用 / 我们为什么需要它？
%s

### 总结您的更改
%s

#### 请注明您已完成以下工作：
- [ ] 确保测试通过，并在需要时添加测试覆盖率。
- [ ] 确保提交消息遵循 [常规提交规范](https://www.conventionalcommits.org/) 的规则。
- [ ] 考虑文档的影响，如果需要，打开一个新的文档问题或文档更改的PR。`, purpose, summary)

	return ok(template)
}

// HandleCheckCodeStructure validates project directory structure
func HandleCheckCodeStructure(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rootDir, _ :=getString(args, "root_dir")
	if rootDir == "" {
		rootDir = "."
	}

	var issues []string
	requiredDirs := []string{
		"src/test/benchmark",
		"src/test/demo",
		"src/test/features",
		"src/test/labs",
	}

	for _, dir := range requiredDirs {
		fullPath := filepath.Join(rootDir, dir)
		if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
			issues = append(issues, fmt.Sprintf("missing required directory: %s", dir))

	}

	// Check for extra directories
	testDir := filepath.Join(rootDir, "src/test")
	files, listErr := os.ReadDir(testDir)
	if listErr != nil {
		return err(fmt.Sprintf("error reading test directory: %v", listErr))
}

	for _, file := range files {
		if file.IsDir() {
			dirName := file.Name()
			if dirName != "benchmark" && dirName != "demo" && dirName != "features" && dirName != "labs" {
				issues = append(issues, fmt.Sprintf("extra directory not following structure guidelines: src/test/%s", dirName))

		}
	}

	if len(issues) > 0 {
		return ok("Code structure issues:\n- " + strings.Join(issues, "\n- "))
}

	return ok("Code structure follows all guidelines ✓")
}
}
}