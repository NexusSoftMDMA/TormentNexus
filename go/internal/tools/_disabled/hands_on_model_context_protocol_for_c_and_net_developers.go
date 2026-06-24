package tools

import "context"

func HandleGetBookInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Hands-On Model Context Protocol for C# and .NET Developers - Published by Packt")
}