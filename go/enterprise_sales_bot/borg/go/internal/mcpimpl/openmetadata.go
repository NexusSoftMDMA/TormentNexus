package mcpimpl

import (
	"context"
	"net/http"
)

func HandleListServices_openmetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceName, _ :=getString(args, "serviceName")
	if serviceName == "" {
		return err("serviceName is required")
}

	_, e := http.DefaultClient.Get("http://localhost:8585/api/v1/services/databaseServices/name/" + serviceName)
	if e != nil {
		return err("HTTP error: " + e.Error())
}

	return success("Service listed successfully")
}

func HandleGetService_openmetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceId, _ :=getString(args, "serviceId")
	if serviceId == "" {
		return err("serviceId is required")
}

	_, e := http.DefaultClient.Get("http://localhost:8585/api/v1/services/databaseServices/" + serviceId)
	if e != nil {
		return err("HTTP error: " + e.Error())
}

	return success("Service retrieved successfully")
}