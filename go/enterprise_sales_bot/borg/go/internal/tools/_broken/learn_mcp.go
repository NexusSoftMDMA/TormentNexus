package tools

import "context"

func HandleLearnMCP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "同学"
	}
	return ok("欢迎" + name + "学习 MCP 协议！")
}