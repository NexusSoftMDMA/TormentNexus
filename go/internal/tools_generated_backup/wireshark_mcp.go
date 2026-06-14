package tools

import "context"

func HandleListInterfaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Interfaces: eth0, wlan0, lo")
}

func HandleAnalyzePcap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	return ok("Analysis complete for " + file)
}