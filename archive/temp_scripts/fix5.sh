sed -i 's/if len(inventory.Tools) > dbToolCountBefore {/if len(inventory.Tools) > dbToolCountBefore || len(inventory.Servers) > 1 {/' go/internal/mcp/inventory.go
