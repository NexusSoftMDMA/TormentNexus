package tools

import "context"

func HandleLearnMCP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bookName, _ :=getString(args, "book")
	if bookName == "" {
		bookName = "Learn Model Context Protocol with TypeScript"
	}
	return ok("Book: " + bookName + " - Packt Publishing")
}

func HandleBookInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Learn Model Context Protocol with TypeScript, published by Packt")
}