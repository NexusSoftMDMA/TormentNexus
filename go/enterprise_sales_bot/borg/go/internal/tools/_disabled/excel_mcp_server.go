package tools

import "context"

func HandleReadCell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	sheet, _ :=getString(args, "sheet")
	cell, _ :=getString(args, "cell")
	if file == "" {
		return err("file is required")
}

	if sheet == "" {
		return err("sheet is required")
}

	if cell == "" {
		return err("cell is required")
}

	return success("Read cell " + cell + " from sheet " + sheet + " in file " + file + " (dummy)")
}

func HandleWriteCell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	sheet, _ :=getString(args, "sheet")
	cell, _ :=getString(args, "cell")
	value, _ :=getString(args, "value")
	if file == "" {
		return err("file is required")
}

	if sheet == "" {
		return err("sheet is required")
}

	if cell == "" {
		return err("cell is required")
}

	if value == "" {
		return err("value is required")
}

	return success("Write cell " + cell + " in file " + file + " with value " + value + " (dummy)")
}