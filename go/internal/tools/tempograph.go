//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
)

func convertTemperature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	value := getFloat(args, "value")
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")

	converted, e := temperatureConvert(value, from, to)
	if e != nil {
		return err(e.Error())
}

	result, _ := json.Marshal(map[string]float64{"result": converted})
	return success(string(result))
}

func listUnits(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	units := []string{"celsius", "fahrenheit", "kelvin"}
	result, _ := json.Marshal(units)
	return success(string(result))
}

func temperatureConvert(value float64, from, to string) (float64, error) {
	var celsius float64
	switch from {
	case "celsius":
		celsius = value
	case "fahrenheit":
		celsius = (value - 32) * 5 / 9
	case "kelvin":
		celsius = value - 273.15
	default:
		return 0, err("unknown from unit")
	switch to {
	case "celsius":
		return celsius, nil
	case "fahrenheit":
		return celsius*9/5 + 32, nil
	case "kelvin":
		return celsius + 273.15, nil
	default:
		return 0, err("unknown to unit")

}

func getFloat(args map[string]interface{}, key string) float64 {
	if v, found := args[key]; found {
		if f, found := v.(float64); found {
			return f
				return 0,
}
}
}
}
}
