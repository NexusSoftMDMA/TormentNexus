package tools

import (
	"bytes"
	"context"
	"os/exec"
)

func HandleFfmpegVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err("failed to get ffmpeg version: " + e.Error())
}

	return success(out.String())
}

func HandleFfmpegProbe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filepath, _ :=getString(args, "filepath")
	if filepath == "" {
		return err("filepath argument is required")
}

	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filepath)
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err("ffprobe failed: " + e.Error())
}

	return ok(out.String())
}