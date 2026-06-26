package tools

import (
	"context"
	"net/url"
	"strings"
)

func HandleDropBeacon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "target_url")
	if targetURL == "" {
		return err("target_url parameter is required")
}

	// Validate URL format
	parsedURL, parseErr := url.Parse(targetURL)
	if parseErr != nil {
		return err("invalid target_url format")
}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return err("target_url must be an HTTP or HTTPS URL")
}

	// In a real implementation, this would actually drop a beacon to the target
	// For this example, we'll just simulate success
	return ok("Beacon dropped to " + targetURL)
}

func HandleListBeacons(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Simulate listing active beacons
	activeBeacons := []string{
		"https://example.com/beacon1",
		"https://example.org/beacon2",
	}

	if len(activeBeacons) == 0 {
		return ok("No active beacons found")
}

	return ok("Active beacons:\n- " + strings.Join(activeBeacons, "\n- "))
}

func HandleClearBeacons(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Simulate clearing all beacons
	return ok("All beacons cleared")
}

func HandleBeaconStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	beaconID, _ :=getString(args, "beacon_id")
	if beaconID == "" {
		return err("beacon_id parameter is required")
}

	// Simulate checking beacon status
	status := "active"
	lastSeen := "2023-11-15T14:30:00Z"

	return ok("Beacon " + beaconID + " status: " + status + ", last seen: " + lastSeen)
}