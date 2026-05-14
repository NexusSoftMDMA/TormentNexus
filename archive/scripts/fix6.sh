sed -i 's/if len(inventory.Tools) > dbToolCountBefore || len(inventory.Servers) > 1 {/inventory.Source = "database"/' go/internal/mcp/inventory.go
