package tools

import (
	"fmt"
	"github.com/NexusSoftMDMA/TormentNexus/mcp"
)

func (r *Registry) registerAdvancedTools() {
	// Provide access to the MCP server manager
	manager := mcp.NewServerManager()

	r.Tools = append(r.Tools, Tool{
		Name:        "install_mcp_server",
		Description: "Installs an MCP server using npx. Arguments: package_name (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			pkg, ok := args["package_name"].(string)
			if !ok {
				return "", fmt.Errorf("package_name must be a string")
			}
			err := manager.InstallNPXServer(pkg)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Successfully installed MCP server: %s", pkg), nil
		},
	})
}
