//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleListLibraries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("AntV libraries: G2, G6, L7, F2, Graphin, AVA. Use GetDocumentation for details.")
}

func HandleGetDocumentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lib, _ :=getString(args, "library")
	switch lib {
	case "G2":
		return success("G2: A grammar of graphics for data visualization. Example: chart.line().data(data).encode('x','year').encode('y','sales').render()
	case "G6":
		return success("G6: Graph visualization engine. Example: new Graph({container:'mount', data, layout:{type:'force'}})"),
	default:
		return err("Unknown library: " + lib + ". Available: G2, G6, L7, F2, Graphin, AVA")

}
}