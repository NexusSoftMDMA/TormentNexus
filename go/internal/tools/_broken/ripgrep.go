package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

func HandleRipgrep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, found := getString(args, "pattern")
	if !found {
		return err("pattern is required")
	}

	dir, found := getString(args, "dir")
	if !found {
		return err("dir is required")
	}

	cmd := exec.CommandContext(ctx, "rg", "--color=never", pattern, dir)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("failed to execute ripgrep: %v", e))
	}

	lines := strings.Split(outb.String(), "\n")
	var results []string
	for _, line := range lines {
		if line != "" {
			results = append(results, line)
		}
	}

	return ok(strings.Join(results, "\n"))
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, found := getString(args, "query")
	if !found {
		return err("query is required")
	}

	urlStr, found := getString(args, "url")
	if !found {
		return err("url is required")
	}

	u, e := url.Parse(urlStr)
	if e != nil {
		return err(fmt.Sprintf("invalid URL: %v", e))
	}

	client := http.DefaultClient

	resp, e := client.Get(u.String())
	if e != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", e))
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response body: %v", e))
	}

	re, e := regexp.Compile(query)
	if e != nil {
		return err(fmt.Sprintf("invalid regex pattern: %v", e))
	}
	matches := re.FindAllString(string(body), -1)

	return ok(strings.Join(matches, "\n"))
}

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, found := getString(args, "dir")
	if !found {
		return err("dir is required")
	}

	files, e := os.ReadDir(dir)
	if e != nil {
		return err(fmt.Sprintf("failed to list files: %v", e))
	}

	var sortedNames []string
	for _, file := range files {
		sortedNames = append(sortedNames, file.Name())
	}

	sort.Strings(sortedNames)

	return ok(strings.Join(sortedNames, "\n"))
}

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, found := getString(args, "file")
	if !found {
		return err("file is required")
	}

	fileInfo, e := os.Stat(filePath)
	if e != nil {
		return err(fmt.Sprintf("failed to get file info: %v", e))
	}

	infoStr := fmt.Sprintf("Size: %d\nModTime: %s\nMode: %s\n", fileInfo.Size(), fileInfo.ModTime(), fileInfo.Mode())

	return ok(infoStr)
}

func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version := "1.0.0"
	return ok(fmt.Sprintf("Version: %s", version))
}