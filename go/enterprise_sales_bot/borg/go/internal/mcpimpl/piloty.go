package mcpimpl

import (
	"context"
)

func HandlePiloty(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Piloty"
	}
	return ok("Hello from " + name + " MCP server!")
}

func HandleListFlights(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Available flights: [{'id':1,'dest':'LAX'},{'id':2,'dest':'JFK'}]")
}