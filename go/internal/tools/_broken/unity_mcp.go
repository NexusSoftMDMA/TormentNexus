package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func HandleBuildCLI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	link, _ :=getBool(args, "link")
	dir, _ :=getString(args, "dir")
	if dir == "" {
		dir = "."
	}

	cmd := exec.CommandContext(ctx, "npm", "install")
	cmd.Dir = dir
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("npm install failed: %v\n%s", e, string(output)))
}

	if link {
		cmd = exec.CommandContext(ctx, "npm", "link")
		cmd.Dir = dir
		output, e = cmd.CombinedOutput()
		if e != nil {
			return err(fmt.Sprintf("npm link failed: %v\n%s", e, string(output)))

	}

	cmd = exec.CommandContext(ctx, "npm", "run", "build")
	cmd.Dir = dir
	output, e = cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("npm run build failed: %v\n%s", e, string(output)))
}

	return ok("Build completed successfully")
}
}