package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListLibvirtDomains(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success(fmt.Sprintf("Libvirt domains listed for %s", name))
}

func HandleGetLibvirtDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uuid, _ :=getString(args, "uuid")
	return success(fmt.Sprintf("Libvirt domain %s info", uuid))
}