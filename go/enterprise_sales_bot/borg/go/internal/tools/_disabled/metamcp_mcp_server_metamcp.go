package tools

import "context"

var mcpServers = make(map[string]string)

func HandleListMcps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	names := make([]string, 0, len(mcpServers))
	for name := range mcpServers {
		names = append(names, name)

	return ok(names)
}

}

func HandleAddMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	url, _ :=getString(args, "url")
	if name == "" || url == "" {
		return err("name and url required")
}

	mcpServers[name] = url
	return success("Mcp added")
}