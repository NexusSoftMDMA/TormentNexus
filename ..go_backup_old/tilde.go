package tools

import (
	"context"
	"os/user"
	"strings"
)

func HandleExpandTilde(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" || !strings.HasPrefix(path, "~") {
		return err("path is missing or does not start with ~")
}

	usr, e := user.Current()
	if e != nil {
		return err("unable to get home directory: " + e.Error())
}

	expanded := strings.Replace(path, "~", usr.HomeDir, 1)
	return success(expanded)
}

func HandleHomeDir(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	usr, e := user.Current()
	if e != nil {
		return err("unable to get home directory: " + e.Error())
}

	return success(usr.HomeDir)
}