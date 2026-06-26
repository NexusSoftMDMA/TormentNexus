package mcpimpl

import (
	"context"
)

func HandleListEngines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	engines := []string{"engine1", "engine2", "engine3"}
	return ok("Available engines: " + join(engines))
}

func HandleRunTuning(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	engine, _ :=getString(args, "engine")
	param, _ :=getString(args, "param")
	if engine == "" {
		return err("engine parameter required")
}

	return success("Tuning started on engine " + engine + " with param " + param)
}

func join_tuning_engines_cli(s []string) string {
	var r string
	for i, v := range s {
		if i > 0 {
			r += ", "
		}
		r += v
	}
	return r
}