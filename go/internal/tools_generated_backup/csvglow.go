package tools

import (
	"context"
	"encoding/csv"
	"strings"
)

func HandleCsvGlowValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	csvStr, _ :=getString(args, "csv_string")
	if csvStr == "" {
		return err("csv_string is required")
}

	reader := csv.NewReader(strings.NewReader(csvStr))
	records, e := reader.ReadAll()
	if e != nil {
		return err("invalid CSV: " + e.Error())
}

	return ok("rows: " + strings.Join([]string{string(rune(len(records)))}, ""))
}