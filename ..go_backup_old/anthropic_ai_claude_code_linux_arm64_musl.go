package tools

import (
	"context"
)

func HandleGetBinaryInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Claude Code native binary for linux-arm64-musl. Architecture: arm64, libc: musl. Latest release available on GitHub.")
}