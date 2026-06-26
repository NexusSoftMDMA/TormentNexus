package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

func HandleAwsCallerIdentity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity")
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err("failed to run aws cli: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(out.Bytes(), &result); e != nil {
		return err("failed to parse output: " + e.Error())
}

	return ok(fmt.Sprintf("AWS caller identity: %v", result))
}