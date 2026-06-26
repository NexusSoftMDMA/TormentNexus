package mcpimpl

import "context"

func HandleGetGreenhouseStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sensorID, _ :=getString(args, "sensorId")
	if sensorID == "" {
		sensorID = "main"
	}
	return ok("Greenhouse " + sensorID + " status: temperature 22°C, humidity 60%")
}

func HandleSetTargetTemperature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getInt(args, "temperature")
	if target == 0 {
		return err("temperature must be set")
}

	return success("target temperature set to " + string(rune(target)) + "°C")
}