sed -i 's/if len(inventory.Tools) > dbToolCountBefore {/inventory.Source = "database"\n\t\tif len(inventory.Tools) > dbToolCountBefore {/' go/internal/mcp/inventory.go
