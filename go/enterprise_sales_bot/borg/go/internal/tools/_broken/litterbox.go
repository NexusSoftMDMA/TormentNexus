package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const litterboxBaseURL = "http://127.0.0.1:1337"

func litterboxGet(path string) (map[string]interface{}, error) {
	client := http.DefaultClient
	resp, getErr := client.Get(litterboxBaseURL + path)
	if getErr != nil {
		return nil, fmt.Errorf("request failed: %w", getErr)
}

	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read body failed: %w", readErr)
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return nil, fmt.Errorf("json decode failed: %w", jsonErr)
}

	return result, nil
}

func litterboxPost(path string, contentType string, body io.Reader) (map[string]interface{}, error) {
	client := http.DefaultClient
	req, reqErr := http.NewRequest("POST", litterboxBaseURL+path, body)
	if reqErr != nil {
		return nil, fmt.Errorf("create request failed: %w", reqErr)
}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)

	resp, postErr := client.Do(req)
	if postErr != nil {
		return nil, fmt.Errorf("request failed: %w", postErr)
}

	defer resp.Body.Close()
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read body failed: %w", readErr)
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(respBody, &result); jsonErr != nil {
		return nil, fmt.Errorf("json decode failed: %w", jsonErr)
}

	return result, nil
}

}

func litterboxGetString(path string) (string, error) {
	client := http.DefaultClient
	resp, getErr := client.Get(litterboxBaseURL + path)
	if getErr != nil {
		return "", fmt.Errorf("request failed: %w", getErr)
}

	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("read body failed: %w", readErr)
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}

	return string(body), nil
}

// HandleUploadPayload uploads a payload file (.exe/.dll/.bin/.lnk/.docx/.xlsx) for analysis.
func HandleUploadPayload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	name, _ :=getString(args, "name")
	if path == "" {
		return err("path is required")
}

	form := url.Values{}
	form.Set("path", path)
	if name != "" {
		form.Set("name", name)

	result, apiErr := litterboxPost("/api/upload", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if apiErr != nil {
		return err(apiErr.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

}

// HandleAnalyze runs static, dynamic, or holygrail analysis on an uploaded file.
func HandleAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileHash, _ :=getString(args, "file_hash")
	analysisType, _ :=getString(args, "analysis_type")
	wait, _ :=getBool(args, "wait")
	cmdArgs, _ :=getString(args, "cmd_args")
	if fileHash == "" {
		return err("file_hash is required")
}

	if analysisType == "" {
		analysisType = "static"
	}
	u := fmt.Sprintf("/api/analyze/%s/%s", url.PathEscape(analysisType), url.PathEscape(fileHash))
	params := url.Values{}
	if wait {
		params.Set("wait", "true")

	if cmdArgs != "" {
		params.Set("cmd_args", cmdArgs)

	if len(params) > 0 {
		u += "?" + params.Encode()

	result, apiErr := litterboxPost(u, "application/json", nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

}
}
}

// HandleGetResults retrieves analysis results (static, dynamic, holygrail, file_info, risk, or comprehensive).
func HandleGetResults(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	resultType, _ :=getString(args, "result_type")
	if target == "" {
		return err("target is required")
}

	if resultType == "" {
		resultType = "comprehensive"
	}
	var path string
	switch resultType {
	case "static":
		path = fmt.Sprintf("/api/results/static/%s", url.PathEscape(target))
	case "dynamic":
		path = fmt.Sprintf("/api/results/dynamic/%s", url.PathEscape(target))
	case "holygrail":
		path = fmt.Sprintf("/api/results/holygrail/%s", url.PathEscape(target))
	case "file_info":
		path = fmt.Sprintf("/api/files/%s", url.PathEscape(target))
	case "risk":
		path = fmt.Sprintf("/api/results/%s/risk", url.PathEscape(target))
	case "comprehensive":
		path = fmt.Sprintf("/api/results/comprehensive/%s", url.PathEscape(target))
	default:
		return err(fmt.Sprintf("unknown result_type: %s", resultType))
}

	result, apiErr := litterboxGet(path)
	if apiErr != nil {
		return err(apiErr.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleGetReport retrieves the full HTML analysis report for a target.
func HandleGetReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	path := fmt.Sprintf("/api/report/%s", url.PathEscape(target))
	html, apiErr := litterboxGetString(path)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(html)
}

// HandleUploadDriver uploads a kernel driver and optionally runs BYOVD analysis.
func HandleUploadDriver(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	name, _ :=getString(args, "name")
	runHolygrail, _ :=getBool(args, "run_holygrail")
	if path == "" {
		return err("path is required")
}

	form := url.Values{}
	form.Set("path", path)
	form.Set("type", "driver")
	if name != "" {
		form.Set("name", name)

	if runHolygrail {
		form.Set("run_holygrail", "true")

	result, apiErr := litterboxPost("/api/upload", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if apiErr != nil {
		return err(apiErr.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

}
}

// HandleValidatePID confirms a PID exists and is accessible before dynamic analysis.
func HandleValidatePID(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pidVal, _ :=getInt(args, "pid")
	if pidVal == 0 {
		return err("pid is required and must be non-zero")
}

	path := fmt.Sprintf("/api/validate/pid/%s", strconv.Itoa(pidVal))
	result, apiErr := litterboxGet(path)
	if apiErr != nil {
		return err(apiErr.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}