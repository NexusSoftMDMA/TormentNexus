package mcpimpl

import "context"

func HandleTranscribe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	audioData, _ :=getString(args, "audio_data")
	if audioData == "" {
		return err("audio_data is required")
}

	return ok("Transcription complete: This is a sample transcription.")
}

func HandleGetModels_mcp_transcribe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available models: whisper-1")
}