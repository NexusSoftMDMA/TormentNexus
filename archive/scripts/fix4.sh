sed -i 's/if len(inventory.Servers) > len(configServers) {/if len(inventory.Servers) > len(configServers) || len(inventory.Servers) > 0 {/' go/internal/mcp/inventory.go
