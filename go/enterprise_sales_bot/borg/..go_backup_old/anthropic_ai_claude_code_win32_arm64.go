package tools

import (
	"context"
)

func HandleGetPlatformInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	platform, _ :=getString(args, "platform")
	if platform == "" {
		return ok("win32-arm64")
}

	if platform == "win32-arm64" {
		return ok("compatible")
}

	return err("incompatible platform: " + platform)
}

func HandleCheckArchitecture(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	arch, _ :=getString(args, "arch")
	if arch == "arm64" {
		return ok("architecture is arm64, compatible")
}

	if arch == "" {
		return ok("no architecture specified, default to arm64")
}

	return err("unexpected architecture: " + arch)
}