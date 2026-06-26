package tools

import (
	"context"
	"encoding/json"
)

func HandleGetGridOptions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	options := []string{
		"columnDefs", "rowData", "pagination", "sortable", "filter",
		"resizable", "editable", "selection", "cellRenderer", "cellEditor",
	}
	data, _ := json.Marshal(options)
	return ok(string(data))
}

func HandleGetOptionDetail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	optionName, _ :=getString(args, "option_name")
	detail := map[string]string{
		"columnDefs": "Defines columns in the grid.",
		"rowData":    "Array of data rows.",
		"pagination": "Enable pagination.",
	}
	if desc, found := detail[optionName]; found {
		return ok(desc)
}

	return err("Option not found")
}