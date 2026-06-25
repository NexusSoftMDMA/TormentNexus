package tools

import "context"

// HandleA2VConvert handles a2v audio-to-video conversion requests.
func HandleA2VConvert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	inputPath, _ := getString(args, "input_path")
	if inputPath == "" {
		return err("input_path is required")
	}
	return ok("a2v conversion not yet implemented")
}