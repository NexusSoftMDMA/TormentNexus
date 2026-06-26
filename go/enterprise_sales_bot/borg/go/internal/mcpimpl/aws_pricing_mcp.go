package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleGetPricing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	service, _ :=getString(args, "service")
	region, _ :=getString(args, "region")
	if service == "" {
		return err("service argument required")
}

	pricing := map[string]float64{
		"ec2": 0.10,
		"s3":  0.023,
		"rds": 0.15,
	}
	price, found := pricing[service]
	if !found {
		return err("unknown service: " + service)
}

	result := map[string]interface{}{
		"service":  service,
		"region":   region,
		"price":    price,
		"currency": "USD",
	}
	data, e := json.Marshal(result)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(data))
}

func HandleListServices_aws_pricing_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	services := []string{"ec2", "s3", "rds"}
	data, e := json.Marshal(services)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(data))
}