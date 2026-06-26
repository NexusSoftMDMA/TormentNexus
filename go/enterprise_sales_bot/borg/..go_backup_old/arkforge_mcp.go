package tools

import (
	"context"
)

func HandleGetServerInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		server = "default"
	}
	return success("Arkforge server: " + server + " is online")
}