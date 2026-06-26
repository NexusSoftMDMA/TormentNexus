package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

func HandleListS3Buckets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	out, e := exec.Command("aws", "s3", "ls", "--output", "json").Output()
	if e != nil {
		return err(fmt.Sprintf("failed to list S3 buckets: %v", e))
}

	var buckets []map[string]interface{}
	if e := json.Unmarshal(out, &buckets); e != nil {
		return err(fmt.Sprintf("failed to parse output: %v", e))
}

	return ok(fmt.Sprintf("Found %d buckets", len(buckets)))
}

func HandleListEC2Instances(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	out, e := exec.Command("aws", "ec2", "describe-instances", "--output", "json").Output()
	if e != nil {
		return err(fmt.Sprintf("failed to list EC2 instances: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(out, &result); e != nil {
		return err(fmt.Sprintf("failed to parse output: %v", e))
}

	reservations, found := result["Reservations"].([]interface{})
	if !found {
		return err("no reservations in response")
}

	return ok(fmt.Sprintf("Found %d reservations", len(reservations)))
}