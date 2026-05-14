sed -i 's/if len(inventory.Tools) > dbToolCountBefore {/if len(inventory.Tools) > dbToolCountBefore || len(inventory.Servers) > len(configServers) {/' go/internal/mcp/inventory.go
