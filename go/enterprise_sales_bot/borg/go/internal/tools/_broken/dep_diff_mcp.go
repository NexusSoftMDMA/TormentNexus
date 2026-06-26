package mcpimpl

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"context"
)

func HandleDepDiff(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file1, _ :=getString(args, "file1")
	file2, _ :=getString(args, "file2")
	if file1 == "" || file2 == "" {
		return err("file1 and file2 are required")
}

	resp1, e := http.DefaultClient.Get(file1)
	if e != nil {
		return err("failed to fetch file1: " + e.Error())
}

	defer resp1.Body.Close()
	body1, e := io.ReadAll(resp1.Body)
	if e != nil {
		return err("failed to read file1: " + e.Error())
}

	resp2, e := http.DefaultClient.Get(file2)
	if e != nil {
		return err("failed to fetch file2: " + e.Error())
}

	defer resp2.Body.Close()
	body2, e := io.ReadAll(resp2.Body)
	if e != nil {
		return err("failed to read file2: " + e.Error())
}

	if bytes.Equal(body1, body2) {
		return ok("No differences found.")
}

	lines1 := strings.Split(string(body1), "\n")
	lines2 := strings.Split(string(body2), "\n")
	var diff strings.Builder
	diff.WriteString("Differences:\n")
	max := len(lines1)
	if len(lines2) > max {
		max = len(lines2)

	for i := 0; i < max; i++ {
		l1 := ""
		if i < len(lines1) {
			l1 = lines1[i]
		}
		l2 := ""
		if i < len(lines2) {
			l2 = lines2[i]
		}
		if l1 != l2 {
			fmt.Fprintf(&diff, "- %s\n+ %s\n", l1, l2)

	}
	return ok(diff.String())
}
}
}