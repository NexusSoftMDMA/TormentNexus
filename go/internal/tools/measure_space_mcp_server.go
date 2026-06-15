//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"strconv"
)

var unitConversions = map[string]float64{

func HandleConvert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	valueStr, _ :=getString(args, "value")
	fromUnit, _ :=getString(args, "from")
	toUnit, _ :=getString(args, "to")
	if fromUnit == "" || toUnit == "" || valueStr == "" {
		return err("missing required arguments: value, from, to")
}

	value, e := strconv.ParseFloat(valueStr, 64)
	if e != nil {
		return err("invalid value: " + e.Error())
}

	fromFactor, found := unitConversions[fromUnit]
	if !found {
		return err("unknown unit: " + fromUnit)
}

	toFactor, found := unitConversions[toUnit]
	if !found {
		return err("unknown unit: " + toUnit)
}

	result := value * fromFactor / toFactor
	return ok(fmt.Sprintf("%.6f %s", result, toUnit))
}

func HandleListUnits(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	units := make([]string, 0, len(unitConversions))
	for u := range unitConversions {
		units = append(units, u)

	return ok(fmt.Sprintf("Available units: %v", units))
}
}
}