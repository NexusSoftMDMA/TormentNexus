package tools

import "context"

func HandleGetResidueInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	residue, _ :=getString(args, "residue")
	if residue == "" {
		return err("residue argument is required")
}

	return ok(`{"residue":"` + residue + `","info":"Standard amino acid"}`)
}

func HandleGetPyRosettaVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version != "" {
		return success("PyRosetta version: " + version)
}

	return success("PyRosetta version: 2024.08")
}