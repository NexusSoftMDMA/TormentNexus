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
	"time"
)

// HandleHttpGet performs a simple HTTP GET request.
// Required argument: "url" (string). Optional: "params" (map[string]interface{}) for query parameters.
func HandleHttpGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rawURL, _ :=getString(args, "url")
	if rawURL == "" {
		return err("missing url")
}

	// Build query parameters if provided
	queryMap, okParams := args["params"].(map[string]interface{})
	if okParams && len(queryMap) > 0 {
		u, parseErr := url.Parse(rawURL)
		if parseErr != nil {
			return err(parseErr.Error())
}

		q := u.Query()
		for k, v := range queryMap {
			q.Set(k, fmt.Sprintf("%v", v))

		u.RawQuery = q.Encode()
		rawURL = u.String()

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

}
}

// HandleFileRead reads the content of a file.
// Required argument: "path" (string). The path can be absolute or relative.
func HandleFileRead(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing path")
}

	absPath, absErr := filepath.Abs(path)
	if absErr != nil {
		return err(absErr.Error())
}

	data, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(data))
}

// HandleExec runs an external command.
// Required argument: "cmd" (string). Optional: "args" ([]interface{}) for command arguments.
func HandleExec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "cmd")
	if cmdStr == "" {
		return err("missing cmd")
}

	var cmdArgs []string
	if rawArgs, found := args["args"].([]interface{}); found {
		for _, a := range rawArgs {
			cmdArgs = append(cmdArgs, fmt.Sprintf("%v", a))

	}
	cmd := exec.CommandContext(ctx, cmdStr, cmdArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(string(output))
}

}

// HandleRegexMatch checks whether the input string matches the given regular expression.
// Required arguments: "pattern" (string), "input" (string).
func HandleRegexMatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		return err("missing pattern")
}

	input, _ :=getString(args, "input")
	if input == "" {
		return err("missing input")
}

	re := regexp.MustCompile(pattern)
	matched := re.MatchString(input)
	return ok(strconv.FormatBool(matched))
}

// HandleSortNumbers sorts a slice of numbers in ascending order.
// Required argument: "numbers" ([]interface{}) containing numeric values.
func HandleSortNumbers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	raw, found := args["numbers"].([]interface{})
	if !ok || len(raw) == 0 {
		return err("missing or invalid numbers")
}

	nums := make([]float64, 0, len(raw))
	for _, v := range raw {
		switch n := v.(type) {
		case float64:
			nums = append(nums, n)
		case float32:
			nums = append(nums, float64(n))
		case int:
			nums = append(nums, float64(n))
		case int64:
			nums = append(nums, float64(n))
		case string:
			if f, convErr := strconv.ParseFloat(n, 64); convErr == nil {
				nums = append(nums, f)
			} else {
				return err("non-numeric value in numbers")
}

		default:
			return err("unsupported number type")

	}
	sort.Float64s(nums)
	// Convert back to []interface{} for JSON marshaling
	sorted := make([]interface{}, len(nums))
	for i, v := range nums {
		sorted[i] = v
	}
	jsonBytes, jsonErr := json.Marshal(sorted)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}
}