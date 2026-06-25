package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	http.DefaultClient = http.DefaultClient
	versionRE  = regexp.MustCompile(`\d+\.\d+\.\d+`)
)

// HandleGetRepositoryInfo retrieves basic information about the Solon repository
func HandleGetRepositoryInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoURL, _ :=getString(args, "repository_url")
	if repoURL == "" {
		repoURL = "https://github.com/opensolon/solon"
	}

	req, reqErr := http.NewRequestWithContext(ctx, "GET", repoURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch repository info: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	// Extract basic info from README
	readmeContent := string(body)
	version := "unknown"
	if matches := versionRE.FindAllString(readmeContent, -1); len(matches) > 0 {
		version = matches[0]
	}

	return ok(fmt.Sprintf("Repository: %s\nLatest Version: %s\nDescription: Java enterprise application development framework", repoURL, version))
}

// HandleListIssueTemplates returns available issue templates in the Solon repository
func HandleListIssueTemplates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	templates := []string{
		"bug_report.md - For reporting bugs",
		"feature_request.md - For suggesting new features",
		"problem_support.md - For asking questions",
		"ISSUE_TEMPLATE.zh-CN.md - Chinese language issue template",
	}

	return ok("Available Issue Templates:\n" + strings.Join(templates, "\n"))
}

// HandleGetTemplateContent retrieves the content of a specific template
func HandleGetTemplateContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	templateName, _ :=getString(args, "template_name")
	if templateName == "" {
		return err("template_name is required")
}

	templateMap := map[string]string{
		"bug_report.md": `### 问题描述
*简要描述您碰到的问题。*

### 如何复现
*请详细告诉我们如何复现您遇到的问题*
1.
2.
3.

### 预期结果
*请告诉我们您预期会发生什么。*

### 实际结果
*请告诉我们实际发生了什么。*`,

		"feature_request.md": `### 请描述您的需求或者改进建议
*对您想要需求或建议的清晰简洁的描述。*

### 请描述你建议的实现方案
*对您想要需求或建议的实现方案的详细描述。*`,

		"problem_support.md": `### 请描述您的问题
*询问有关本项目的使用和其他方面的相关问题。*`,

		"ISSUE_TEMPLATE.zh-CN.md": `### 问题描述
*请详细描述您遇到的问题*

### 我当前使用 Solon 版本是?
*请填写您使用的版本号*`,
	}

	content, exists := templateMap[templateName]
	if !exists {
		return err(fmt.Sprintf("template '%s' not found", templateName))
}

	return ok(fmt.Sprintf("Template: %s\n\n%s", templateName, content))
}

// HandleListContributionGuidelines returns the contribution guidelines
func HandleListContributionGuidelines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	guidelines := []string{
		"1. 版权说明: 本仓库的源码版权归 noear 开源组织所有",
		"2. 贡献分类: 代码贡献、合作贡献、其他贡献",
		"3. 代码贡献流程:",
		"   - 提交 Issue 并确认",
		"   - Fork 仓库",
		"   - 在 main 分支上编写代码",
		"   - 添加单元测试",
		"   - 提交 PR 到 main 分支",
		"4. 代码分支保护规则: main 分支禁止直接推送",
		"5. 提交信息规范: 遵循常规提交规范",
	}

	return ok("Solon Contribution Guidelines:\n" + strings.Join(guidelines, "\n"))
}

// HandleGetCodeRepositoryList returns the list of Solon code repositories
func HandleGetCodeRepositoryList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repositories := []map[string]string{
		{"name": "solon", "url": "https://gitee.com/opensolon/solon", "description": "主代码仓库"},
		{"name": "solon-examples", "url": "https://gitee.com/opensolon/solon-examples", "description": "官网配套示例代码仓库"},
		{"name": "solon-ai", "url": "https://gitee.com/opensolon/solon-ai", "description": "AI 相关代码仓库"},
		{"name": "solon-cloud", "url": "https://gitee.com/opensolon/solon-cloud", "description": "云相关代码仓库"},
		{"name": "solon-admin", "url": "https://gitee.com/opensolon/solon-admin", "description": "管理相关代码仓库"},
		{"name": "solon-java17", "url": "https://gitee.com/opensolon/solon-java17", "description": "Java17 版本代码仓库"},
		{"name": "solon-java25", "url": "https://gitee.com/opensolon/solon-java25", "description": "Java25 版本代码仓库"},
	}

	var repoList []string
	for _, repo := range repositories {
		repoList = append(repoList, fmt.Sprintf("%s: %s (%s)", repo["name"], repo["url"], repo["description"]))

	sort.Strings(repoList)
	return ok("Solon Code Repositories:\n" + strings.Join(repoList, "\n"))
}

}

// HandleGetPRTemplateContent returns the pull request template content
func HandleGetPRTemplateContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prTemplate := `### 这个PR有什么用 / 我们为什么需要它？
*请说明这个PR的目的和必要性*

### 总结您的更改
*请总结您所做的更改*

#### 请注明您已完成以下工作：
- [ ] 确保测试通过，并在需要时添加测试覆盖率。
- [ ] 确保提交消息遵循 [常规提交规范](https://www.conventionalcommits.org/) 的规则。
- [ ] 考虑文档的影响，如果需要，打开一个新的文档问题或文档更改的PR。`

	return ok("Pull Request Template:\n\n" + prTemplate)
}