package tools

import "context"

func HandleRtcStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("RTC transport mode: server")
}

func HandleRtcConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	peerId, _ :=getString(args, "peerId")
	return success("Connected to peer: " + peerId)
}