package mcpimpl

import "context"

func HandleGetGenomeInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	genome, _ :=getString(args, "genome")
	return ok("Genome info for " + genome)
}

func HandleListAssemblies_ucsc_genome_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	species, _ :=getString(args, "species")
	return success("Assemblies for " + species)
}