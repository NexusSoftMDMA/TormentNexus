package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func HandleCreateMcpServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "my-mcp-server"
	}
	dir := name

	e := os.MkdirAll(dir, 0755)
	if e != nil {
		return err("failed to create directory: " + e.Error())
}

	packageJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "^0.5.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}`, name)
	e = os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
	if e != nil {
		return err("failed to write package.json: " + e.Error())
}

	tsconfig := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ES2020",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src/**/*"]
}`
	e = os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(tsconfig), 0644)
	if e != nil {
		return err("failed to write tsconfig.json: " + e.Error())
}

	e = os.MkdirAll(filepath.Join(dir, "src"), 0755)
	if e != nil {
		return err("failed to create src directory: " + e.Error())
}

	mainTS := `import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

const server = new Server(
  { name: "example", version: "1.0.0" },
  { capabilities: { tools: {} } }
);

const transport = new StdioServerTransport();
await server.connect(transport);
`
	e = os.WriteFile(filepath.Join(dir, "src", "index.ts"), []byte(mainTS), 0644)
	if e != nil {
		return err("failed to write src/index.ts: " + e.Error())
}

	return ok("Scaffolded MCP server project at " + dir)
}