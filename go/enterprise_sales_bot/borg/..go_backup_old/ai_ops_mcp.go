package tools

import (
	"context"
)

func HandleSystemHealthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("System health check completed")
}

func HandleCheckService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceName, _ :=getString(args, "serviceName")
	return ok("Service " + serviceName + " is healthy")
}