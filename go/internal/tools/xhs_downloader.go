package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// shared HTTP client
var http.DefaultClient = http.DefaultClient

// HandleExtract extracts information from a XiaoHongShu link.
// Expected args:
//   - "url" (string, required): the XHS link.
//   - "download" (bool, optional, default false): whether to download the media.
//   - "index" (array of numbers, optional): image indices for image‑only notes.
func HandleExtract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if strings.TrimSpace(urlStr) == "" {
		return err("parameter \"url\" is required")
}

	download, _ :=getBool(args, "download")

	// Simple validation of the URL format.
	if !strings.HasPrefix(urlStr, "http") {
		return err("invalid url format")
}

	// In a real implementation we would call the XHS API here.
	// For the MCP stub we just return a formatted message.
	msg := fmt.Sprintf("Extracted XHS data for %s (download=%v)", urlStr, download)
	return ok(msg)
}

// HandleDownload downloads the media pointed to by a XHS link.
// Expected args:
//   - "url" (string, required): the XHS link.
//   - "dest" (string, optional): destination directory; defaults to current working directory.
func HandleDownload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if strings.TrimSpace(urlStr) == "" {
		return err("parameter \"url\" is required")
}

	destDir, _ :=getString(args, "dest")
	if destDir == "" {
		destDir = "."
	}

	// Resolve absolute path.
	absDest, resolveErr := filepath.Abs(destDir)
	if resolveErr != nil {
		return err("cannot resolve destination path: " + resolveErr.Error())
}

	// Create destination directory if it does not exist.
	mkdirErr := os.MkdirAll(absDest, 0755)
	if mkdirErr != nil {
		return err("failed to create destination directory: " + mkdirErr.Error())
}

	// Perform a GET request.
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if reqErr != nil {
		return err("failed to create request: " + reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err("http request failed: " + fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("unexpected HTTP status: " + strconv.Itoa(resp.StatusCode))
}

	// Derive filename from URL.
	segments := strings.Split(urlStr, "/")
	fileName := segments[len(segments)-1]
	if fileName == "" {
		fileName = "downloaded_file"
	}
	filePath := filepath.Join(absDest, fileName)

	// Write to file.
	outFile, fileErr := os.Create(filePath)
	if fileErr != nil {
		return err("cannot create file: " + fileErr.Error())
}

	defer outFile.Close()

	_, copyErr := io.Copy(outFile, resp.Body)
	if copyErr != nil {
		return err("failed to write file: " + copyErr.Error())
}

	msg := fmt.Sprintf("Downloaded %s to %s", urlStr, filePath)
	return ok(msg)
}

// HandleInfo returns a short description of the XHS‑Downloader tool.
func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	info := "XHS‑Downloader: extract links, fetch metadata and download media from XiaoHongShu (RedNote)."
	return ok(info)
}

// HandleVersion returns the static version string of this tool module.
func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// In a real project this could be injected at build time.
	const version = "v1.0.0"
	return ok(version)
}