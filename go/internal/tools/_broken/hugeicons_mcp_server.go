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
	"sort"
	"strconv"
	"strings"
	"time"
)

const hugeiconsAPIBase = "https://hugeicons.com/api/v1"

var http.DefaultClient = http.DefaultClient

type Icon struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Style       string   `json:"style"`
	PreviewURL  string   `json:"preview_url"`
	DownloadURL string   `json:"download_url"`
}

type IconSearchResponse struct {
	Icons []Icon `json:"icons"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

type IconCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IconCount   int    `json:"icon_count"`
	Description string `json:"description"`
}

type IconStyle struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func HandleSearchIcons(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	category, _ :=getString(args, "category")
	style, _ :=getString(args, "style")
	page, _ :=getInt(args, "page")
	limit, _ :=getInt(args, "limit")

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	apiURL, apiErr := url.Parse(hugeiconsAPIBase + "/icons/search")
	if apiErr != nil {
		return err(apiErr.Error())
}

	q := apiURL.Query()
	if query != "" {
		q.Set("q", query)

	if category != "" {
		q.Set("category", category)

	if style != "" {
		q.Set("style", style)

	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	apiURL.RawQuery = q.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hugeicons-mcp-server/1.0")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var searchResp IconSearchResponse
	parseErr := json.Unmarshal(body, &searchResp)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d icons (page %d, limit %d):\n\n", searchResp.Total, searchResp.Page, searchResp.Limit))
	for _, icon := range searchResp.Icons {
		result.WriteString(fmt.Sprintf("- %s (ID: %s)\n", icon.Name, icon.ID))
		result.WriteString(fmt.Sprintf("  Category: %s | Style: %s\n", icon.Category, icon.Style))
		if len(icon.Tags) > 0 {
			result.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(icon.Tags, ", ")))

		result.WriteString(fmt.Sprintf("  Preview: %s\n", icon.PreviewURL))
		result.WriteString("\n")

	return ok(result.String())
}

}
}
}
}
}

func HandleGetIconDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	iconID, _ :=getString(args, "icon_id")

	if iconID == "" {
		return err("icon_id is required")
}

	apiURL := hugeiconsAPIBase + "/icons/" + url.PathEscape(iconID)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hugeicons-mcp-server/1.0")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var icon Icon
	parseErr := json.Unmarshal(body, &icon)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Icon: %s\n", icon.Name))
	result.WriteString(fmt.Sprintf("ID: %s\n", icon.ID))
	result.WriteString(fmt.Sprintf("Category: %s\n", icon.Category))
	result.WriteString(fmt.Sprintf("Style: %s\n", icon.Style))
	if len(icon.Tags) > 0 {
		result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(icon.Tags, ", ")))

	result.WriteString(fmt.Sprintf("Preview URL: %s\n", icon.PreviewURL))
	result.WriteString(fmt.Sprintf("Download URL: %s\n", icon.DownloadURL))

	return ok(result.String())
}

}

func HandleListCategories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL := hugeiconsAPIBase + "/categories"

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hugeicons-mcp-server/1.0")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var categories []IconCategory
	parseErr := json.Unmarshal(body, &categories)
	if parseErr != nil {
		return err(parseErr.Error())
}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Available Categories (%d total):\n\n", len(categories)))
	for _, cat := range categories {
		result.WriteString(fmt.Sprintf("- %s (ID: %s)\n", cat.Name, cat.ID))
		result.WriteString(fmt.Sprintf("  Icons: %d\n", cat.IconCount))
		if cat.Description != "" {
			result.WriteString(fmt.Sprintf("  Description: %s\n", cat.Description))

		result.WriteString("\n")

	return ok(result.String())
}

}
}

func HandleListStyles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL := hugeiconsAPIBase + "/styles"

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hugeicons-mcp-server/1.0")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var styles []IconStyle
	parseErr := json.Unmarshal(body, &styles)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Available Styles (%d total):\n\n", len(styles)))
	for _, style := range styles {
		result.WriteString(fmt.Sprintf("- %s (ID: %s)\n", style.Name, style.ID))

	return ok(result.String())
}

}

func HandleDownloadIcon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	iconID, _ :=getString(args, "icon_id")
	format, _ :=getString(args, "format")
	outputPath, _ :=getString(args, "output_path")

	if iconID == "" {
		return err("icon_id is required")
}

	if format == "" {
		format = "svg"
	}
	if outputPath == "" {
		return err("output_path is required")
}

	validFormats := map[string]bool{"svg": true, "png": true, "pdf": true, "json": true}
	if !validFormats[format] {
		return err(fmt.Sprintf("invalid format '%s', must be one of: svg, png, pdf, json", format))
}

	detailsURL := hugeiconsAPIBase + "/icons/" + url.PathEscape(iconID)
	req, reqErr := http.NewRequestWithContext(ctx, "GET", detailsURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hugeicons-mcp-server/1.0")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	body, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var icon Icon
	parseErr := json.Unmarshal(body, &icon)
	if parseErr != nil {
		return err(parseErr.Error())
}

	downloadURL := icon.DownloadURL
	if downloadURL == "" {
		downloadURL = fmt.Sprintf("%s/icons/%s/download?format=%s", hugeiconsAPIBase, url.QueryEscape(iconID), format)
	} else {
		parsedURL, parseErr := url.Parse(downloadURL)
		if parseErr == nil {
			q := parsedURL.Query()
			q.Set("format", format)
			parsedURL.RawQuery = q.Encode()
			downloadURL = parsedURL.String()

	}

	downloadReq, dlReqErr := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if dlReqErr != nil {
		return err(dlReqErr.Error())
}

	downloadReq.Header.Set("User-Agent", "hugeicons-mcp-server/1.0")

	dlResp, dlFetchErr := http.DefaultClient.Do(downloadReq)
	if dlFetchErr != nil {
		return err(dlFetchErr.Error())
}

	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		dlBody, dlReadErr := io.ReadAll(dlResp.Body)
		if dlReadErr != nil {
			return err(dlReadErr.Error())
}

		return err(fmt.Sprintf("download failed with status %d: %s", dlResp.StatusCode, string(dlBody)))
}

	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		mkdirErr := os.MkdirAll(dir, 0755)
		if mkdirErr != nil {
			return err(mkdirErr.Error())

	}

	file, fileErr := os.Create(outputPath)
	if fileErr != nil {
		return err(fileErr.Error())
}

	_, copyErr := io.Copy(file, dlResp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		return err(copyErr.Error())
}

	if closeErr != nil {
		return err(closeErr.Error())
}

	return ok(fmt.Sprintf("Successfully downloaded icon '%s' (%s) to: %s", icon.Name, format, outputPath))
}

}
}

func HandleConvertIcon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	inputPath, _ :=getString(args, "input_path")
	outputPath, _ :=getString(args, "output_path")
	targetFormat, _ :=getString(args, "target_format")

	if inputPath == "" {
		return err("input_path is required")
}

	if outputPath == "" {
		return err("output_path is required")
}

	if targetFormat == "" {
		return err("target_format is required")
}

	validFormats := map[string]bool{"svg": true, "png": true, "pdf": true, "ico": true}
	if !validFormats[targetFormat] {
		return err(fmt.Sprintf("invalid target_format '%s', must be one of: svg, png, pdf, ico", targetFormat))
}

	_, statErr := os.Stat(inputPath)
	if statErr != nil {
		return err(fmt.Sprintf("input file not found: %s", statErr.Error()))
}

	ext := filepath.Ext(inputPath)
	sourceFormat := strings.TrimPrefix(ext, ".")
	if sourceFormat == "" {
		return err("cannot determine source format from input path")
}

	var cmd *exec.Cmd
	switch {
	case sourceFormat == "svg" && targetFormat == "png":
		_, inkscapeErr := exec.LookPath("inkscape")
		if inkscapeErr == nil {
			cmd = exec.CommandContext(ctx, "inkscape", inputPath, "--export-filename="+outputPath, "--export-dpi=300")
		} else {
			_, cairosvgErr := exec.LookPath("cairosvg")
			if cairosvgErr == nil {
				cmd = exec.CommandContext(ctx, "cairosvg", inputPath, "-o", outputPath)
			} else {
				return err("no SVG to PNG converter found (install inkscape or cairosvg)")

		}
	case sourceFormat == "png" && targetFormat == "svg":
		_, potraceErr := exec.LookPath("potrace")
		if potraceErr == nil {
			cmd = exec.CommandContext(ctx, "potrace", "-s", "-o", outputPath, inputPath)
		} else {
			return err("no PNG to SVG converter found (install potrace)")
}

	case sourceFormat == "svg" && targetFormat == "pdf":
		_, inkscapeErr := exec.LookPath("inkscape")
		if inkscapeErr == nil {
			cmd = exec.CommandContext(ctx, "inkscape", inputPath, "--export-filename="+outputPath)
		} else {
			return err("no SVG to PDF converter found (install inkscape)")
}

	default:
		_, convertErr := exec.LookPath("convert")
		if convertErr == nil {
			cmd = exec.CommandContext(ctx, "convert", inputPath, outputPath)
		} else {
			return err(fmt.Sprintf("conversion from %s to %s not supported (install ImageMagick)", sourceFormat, targetFormat))

	}

	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		mkdirErr := os.MkdirAll(dir, 0755)
		if mkdirErr != nil {
			return err(mkdirErr.Error())

	}

	runErr := cmd.Run()
	if runErr != nil {
		return err(fmt.Sprintf("conversion failed: %s", runErr.Error()))
}

	return ok(fmt.Sprintf("Successfully converted %s to %s: %s", inputPath, targetFormat, outputPath))
}

}
}
}

func HandleGetIconCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	iconID, _ :=getString(args, "icon_id")
	framework, _ :=getString(args, "framework")
	style, _ :=getString(args, "style")

	if iconID == "" {
		return err("icon_id is required")
}

	if framework == "" {
		framework = "react"
	}

	validFrameworks := map[string]bool{
		"react": true, "vue": true, "angular": true,
		"html": true, "flutter": true, "swift": true,
	}
	if !validFrameworks[framework] {
		return err(fmt.Sprintf("invalid framework '%s', must be one of: react, vue, angular, html, flutter, swift", framework))
}

	apiURL := hugeiconsAPIBase + "/icons/" + url.PathEscape(iconID)
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hugeicons-mcp-server/1.0")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var icon Icon
	parseErr := json.Unmarshal(body, &icon)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Icon: %s (ID: %s)\n\n", icon.Name, icon.ID))

	switch framework {
	case "react":
		result.WriteString(fmt.Sprintf("import { %s } from 'hugeicons-react';\n\n", icon.Name))
		result.WriteString(fmt.Sprintf("<%s size=\"24\" color=\"currentColor\" />", icon.Name))
	case "vue":
		result.WriteString(fmt.Sprintf("<template>\n  <HugeIcon :name=\"'%s'\" :size=\"24\" />\n</template>\n\n", icon.ID))
		result.WriteString("import { HugeIcon } from 'hugeicons-vue';")
	case "angular":
		result.WriteString(fmt.Sprintf("<huge-icon name=\"%s\" [size]=\"24\"></huge-icon>\n\n", icon.ID))
		result.WriteString("import { HugeIconModule } from 'hugeicons-angular';")
	case "html":
		result.WriteString(fmt.Sprintf('<script src="https://unpkg.com/hugeicons"></script>\n\n'))
		result.WriteString(fmt.Sprintf('<i class="hgi hgi-%s" style="font-size: 24px;"></i>', icon.ID))
	case "flutter":
		result.WriteString(fmt.Sprintf("import 'package:hugeicons_flutter/hugeicons_flutter.dart';\n\n"))
		result.WriteString(fmt.Sprintf("HugeIcon(icon: HugeIcons.%s, size: 24)", strings.ToLower(icon.ID)))
	case "swift":
		result.WriteString(fmt.Sprintf("import HugeIcons\n\n"))
		result.WriteString(fmt.Sprintf("HugeIcon(.%s, size: 24)", strings.ToLower(icon.ID)))

	if style != "" {
		result.WriteString(fmt.Sprintf("\n\nStyle variant: %s", style))

	return ok(result.String())
}
}
}