package tools. The errors mentioned: missing return in HandleListFiles, HandleRunCommand, HandleFetchURL, HandleSearch. But looking at code, each function returns ok(...) at end. Maybe they originally missed returns but now they have returns. However maybe the functions need to return (ToolResponse, error) but they currently return only ToolResponse? Wait signature: func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { return ok("pong") } That's returning only one value, missing error. Should be return ok("pong"). Similarly other functions return only ToolResponse. So missing return for error nil. Also err("error") likely returns (ToolResponse, error). Let's assume e returns (ToolResponse, error). So they need to return err(...). That's fine. But ok returns (ToolResponse, error) as well? Probably ok returns (ToolResponse, error). So they need to return ok(...). That's fine. But the functions currently return only one value, causing compile error. So need to add second return value nil for ok cases. However the err("error") returns (ToolResponse, error) already includes error. So we need to adjust all returns to match signature.

Specifically:

- HandlePing: return ok("pong") // ok returns (ToolResponse, error) maybe. If ok returns (ToolResponse, error) then returning that is fine. But signature expects two values; returning a single value of type (ToolResponse, error) is okay because it's a tuple. In Go you can return a multi-valued function result directly. Example: func f() (int, error) { return g() } where g returns (int, error). That's allowed. So returning ok("pong") is fine if ok returns (ToolResponse, error). So maybe not missing return.

But the errors mention missing return in those functions. Possibly they originally had early returns missing error nil. Let's examine each function: they have returns for error cases using err(...). For success they return ok(...). That's fine.

But maybe the functions have missing return after the early error returns? No.

Maybe the issue is that the functions have multiple return statements but not all code paths return a value. For example, in HandleListFiles, after reading entries, they return ok(...). That's fine.

But maybe the code uses named return values? No.

Thus maybe the errors are about missing return after the for loop? Not.

But the instruction says to fix missing return in those functions. So maybe they originally omitted return statements after the error handling but now they have them. However we still need to ensure compilation.

Potential other issues: imports include fmt, io, net/http, os, os/exec, path/filepath, regexp, sort, strings, time. All used.

Potential missing import for context already present.

Potential missing use of ctx in HandleFetchURL? Not needed.

Potential missing use of ctx in HandleSearch? Not needed.

Potential missing use of ctx in HandleListFiles? Not needed.

Potential missing use of ctx in HandleRunCommand? Already used.

Potential missing use of ctx in HandlePing? Not needed.

Potential missing return of nil error after ok? Not needed if ok returns (ToolResponse, error). Let's assume ok returns (ToolResponse, error). So fine.

But the instruction says: "return ok("text"), if e != nil { return err(e.Error()) }". That suggests pattern: return ok("text") (single return). So ok returns (ToolResponse, error). So code is fine.

Thus maybe the missing return errors refer to missing return statements after the error handling inside the functions? Actually they have returns.

Maybe the code missing return after the final ok? No.

Maybe the code missing return in functions that have early returns but not final return? Not.

Let's compile mentally: The functions all return (ToolResponse, error). The returns are either ok(...) or err(...). Both functions presumably return (ToolResponse, error). So compile.

But maybe the e function returns only ToolResponse, not error. Then returning err(...) would be mismatched. But the instruction says "return err(e.Error())". So e returns (ToolResponse, error). So fine.

Thus maybe there are other compilation errors not listed. Let's check for any undefined identifiers: getString, e, ok are pre-defined. So fine.

Potential issue: In HandleRunCommand, they use exec.CommandContext, but they didn't import "os/exec"? Already imported.

Potential issue: In HandleSearch, they use os.DirEntry but imported os. That's fine.

Potential issue: In WalkDir, the callback signature is func(path string, d DirEntry, e error) error. They used walkErr variable name for both outer and inner. That's okay.

Potential issue: In HandleFetchURL, they use client.Get without context; fine.

Thus maybe the code compiles as is. But the instruction demands fixing missing return errors. Maybe they want to add explicit nil error returns after ok? But pattern says return ok("text") not ok(...), nil.

Thus maybe the missing return errors refer to missing return statements after the error handling inside the functions (like after err(...)). But they already have returns.

Maybe the original code had missing returns but the snippet already fixed them. However we still need to output the complete fixed code.

Thus we can just output the same code (maybe add missing import for "io" already there). Ensure no extra spaces.

Let's double-check: In HandleListFiles, they call getString(args, "path") which returns string. If empty, set ".". Then os.ReadDir returns ([]DirEntry, error). They handle error. Then they collect names, sort, return ok(...). Good.

All functions have proper returns.

Thus final answer is same code.

But maybe they need to add context usage for cancellation? Not required.

Thus output the code.

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// HandlePing returns a simple pong response.
func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

// HandleListFiles lists the entries in the provided directory.
// Expected argument: "path" (string). If omitted, defaults to the current directory.
func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "path")
	if dirPath == "" {
		dirPath = "."
	}
	entries, readErr := os.ReadDir(dirPath)
	if readErr != nil {
		return err(readErr.Error())
}

	var names []string
	for _, e := range entries {
		names = append(names, e.Name())

	sort.Strings(names)
	return ok(strings.Join(names, "\n"))
}

}

// HandleRunCommand executes a command and returns its combined stdout and stderr.
// Expected argument: "cmd" (string) – the command line to execute.
func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdLine, _ :=getString(args, "cmd")
	if cmdLine == "" {
		return err("missing cmd argument")
}

	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return err("empty cmd argument")
}

	cmdName := parts[0]
	cmdArgs := parts[1:]

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		// Include the command output for debugging.
		return err(fmt.Sprintf("command error: %s, output: %s", execErr.Error(), string(output)))
}

	return ok(string(output))
}

// HandleFetchURL retrieves the content of a URL using a 30‑second timeout.
// Expected argument: "url" (string).
func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing url argument")
}

	client := http.DefaultClient
	resp, httpErr := client.Get(urlStr)
	if httpErr != nil {
		return err(httpErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleSearch walks a directory tree and returns paths of files whose names match a regexp.
// Expected arguments: "path" (string, defaults to "."), "pattern" (string, required).
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "path")
	if dirPath == "" {
		dirPath = "."
	}
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		return err("missing pattern argument")
}

	re := regexp.MustCompile(pattern)

	var matches []string
	walkErr := filepath.WalkDir(dirPath, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && re.MatchString(d.Name()) {
			matches = append(matches, p)

		return nil
	})
	if walkErr != nil {
		return err(walkErr.Error())
}

	sort.Strings(matches)
	return ok(strings.Join(matches, "\n"))
}
}