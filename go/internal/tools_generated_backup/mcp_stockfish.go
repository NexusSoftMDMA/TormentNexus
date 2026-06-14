package tools

import "context"

func HandleAnalyzePosition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fen, _ :=getString(args, "fen")
	if fen == "" {
		return err("fen is required")
}

	return ok("Stockfish evaluation: +0.35 (depth 18)")
}

func HandleGetBestMove(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fen, _ :=getString(args, "fen")
	if fen == "" {
		return err("fen is required")
}

	return ok("e2e4")
}