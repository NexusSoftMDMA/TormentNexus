package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"net/url"
)

// Category represents a category in the Awesome-MCP-ZH list
type mcpCategory struct {
	Name        string
	Emoji       string
	Description string
	Subcategories []string
}

// Entry represents a single MCP resource entry
type mcpEntry struct {
	Name        string
	URL         string
	Description string
	Notes       string
	Category    string
}

// getCategories returns the known categories from Awesome-MCP-ZH
func getCategories() []mcpCategory {
	return []mcpCategory{
}
		{Name: "搜索", Emoji: "🔍", Description: "搜索类 MCP 服务器", Subcategories: []string{"通用搜索", "代码搜索", "学术搜索"}},
		{Name: "数据库", Emoji: "🗄️", Description: "数据库连接 MCP 服务器", Subcategories: []string{"SQL数据库", "NoSQL数据库", "图数据库"}},
		{Name: "文件系统", Emoji: "📁", Description: "文件系统操作 MCP 服务器", Subcategories: []string{"本地文件", "云存储", "文档管理"}},
		{Name: "开发工具", Emoji: "🛠️", Description: "开发工具类 MCP 服务器", Subcategories: []string{"Git", "CI/CD", "代码分析"}},
		{Name: "浏览器自动化", Emoji: "🌐", Description: "浏览器自动化 MCP 服务器", Subcategories: []string{"网页抓取", "自动化测试", "表单填写"}},
		{Name: "通信", Emoji: "💬", Description: "通信与消息类 MCP 服务器", Subcategories: []string{"邮件", "即时通讯", "社交媒体"}},
		{Name: "AI与机器学习", Emoji: "🤖", Description: "AI 与机器学习 MCP 服务器", Subcategories: []string{"模型服务", "向量数据库", "RAG"}},
		{Name: "生产力", Emoji: "📊", Description: "生产力工具 MCP 服务器", Subcategories: []string{"日历", "任务管理", "笔记"}},
		{Name: "监控", Emoji: "📈", Description: "监控与可观测性 MCP 服务器", Subcategories: []string{"日志", "指标", "告警"}},
		{Name: "安全", Emoji: "🔒", Description: "安全相关 MCP 服务器", Subcategories: []string{"认证", "扫描", "合规"}},
	}
}

// HandleListCategories lists all available MCP resource categories from Awesome-MCP-ZH
func HandleListCategories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	categories := getCategories()
	var sb strings.Builder
	sb.WriteString("Awesome-MCP-ZH 收录分类列表:\n\n")
	for i, cat := range categories {
		sb.WriteString(fmt.Sprintf("%d. %s %s — %s\n", i+1, cat.Emoji, cat.Name, cat.Description))
		if len(cat.Subcategories) > 0 {
			sb.WriteString(fmt.Sprintf("   子分类: %s\n", strings.Join(cat.Subcategories, "、")))

	}
	sb.WriteString(fmt.Sprintf("\n共 %d 个分类。使用 search_entries 工具搜索具体条目。\n", len(categories)))
	return ok(sb.String())
}

// HandleSearchEntries searches MCP entries by keyword
func HandleSearchEntries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword 参数不能为空")
}

	entries := getSampleEntries()
	keyword = strings.ToLower(keyword)
	var matched []mcpEntry
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), keyword) ||
			strings.Contains(strings.ToLower(entry.Description), keyword) ||
			strings.Contains(strings.ToLower(entry.Category), keyword) ||
			strings.Contains(strings.ToLower(entry.Notes), keyword) {
			matched = append(matched, entry)

	}

	if len(matched) == 0 {
		return ok(fmt.Sprintf("未找到与「%s」相关的 MCP 条目。请尝试其他关键词。", keyword))
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("搜索「%s」的结果(共 %d 条):\n\n", keyword, len(matched)))
	sb.WriteString("| 名称 | 中文介绍 | 备注 |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")
	for _, e := range matched {
		sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s |\n", e.Name, e.URL, e.Description, e.Notes))

	return ok(sb.String())
}

}
}

// HandleGetContributingGuide returns the contributing guidelines for Awesome-MCP-ZH
func HandleGetContributingGuide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var sb strings.Builder
	sb.WriteString("Awesome-MCP-ZH 贡献指南\n\n")
	// ... (same as original code)
	return ok(sb.String())
}

// HandleFetchReadme fetches the README content from the Awesome-MCP-ZH GitHub repository
func HandleFetchReadme(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := http.DefaultClient
	resp, fetchErr := client.Get("https://raw.githubusercontent.com/yzfly/Awesome-MCP-ZH/main/README.md")
	if fetchErr != nil {
		return err(fmt.Sprintf("获取 README 失败: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("获取 README 失败,HTTP 状态码: %d", resp.StatusCode))
}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("读取响应失败: %s", readErr.Error()))
}

	content := string(bodyBytes)
	// Truncate if too long
	maxLen := 8000
	if len(content) > maxLen {
		content = content[:maxLen] + "\n\n... (内容已截断,完整内容请访问 GitHub 仓库)"
	}

	return ok(content)
}

// HandleValidateEntry checks if a proposed entry meets the contributing guidelines
func HandleValidateEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// ... (same as original code)
	return ok(sb.String())
}

// urlParse wraps url.Parse to avoid variable shadowing
func urlParse(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}