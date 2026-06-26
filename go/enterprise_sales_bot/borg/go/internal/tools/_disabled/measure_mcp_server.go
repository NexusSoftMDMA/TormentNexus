package tools

import (
	"context"
	"fmt"
	"strconv"
)

func HandleConvertLength(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	valueStr, _ :=getString(args, "value")
	unit, _ :=getString(args, "unit")
	value, e := strconv.ParseFloat(valueStr, 64)
	if e != nil {
		return err("invalid value")
}

	var result float64
	switch unit {
	case "m_to_ft":
		result = value * 3.28084
	case "ft_to_m":
		result = value / 3.28084
	case "in_to_cm":
		result = value * 2.54
	case "cm_to_in":
		result = value / 2.54
	default:
		return err("unsupported unit")
}

	return ok(fmt.Sprintf("%.2f", result))
}

func HandleConvertTemperature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	valueStr, _ :=getString(args, "value")
	scale, _ :=getString(args, "scale")
	value, e := strconv.ParseFloat(valueStr, 64)
	if e != nil {
		return err("invalid value")
}

	var result float64
	switch scale {
	case "C_to_F":
		result = value*9/5 + 32
	case "F_to_C":
		result = (value - 32) * 5 / 9
	case "C_to_K":
		result = value + 273.15
	case "K_to_C":
		result = value - 273.15
	default:
		return err("unsupported scale")
}

	return ok(fmt.Sprintf("%.2f", result))
}