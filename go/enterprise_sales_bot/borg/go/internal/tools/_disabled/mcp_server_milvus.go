package tools

import (
	"context"
	"strconv"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection_name")
	query, _ :=getString(args, "query_vector")
	topK, _ :=getInt(args, "top_k")
	if collection == "" || query == "" {
		return err("collection_name and query_vector are required")
	}
	msg := "Searched in " + collection + " for top " + strconv.Itoa(topK) + " results"
	return ok(msg)
}

func HandleInsert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection_name")
	vector, _ :=getString(args, "vector")
	if collection == "" || vector == "" {
		return err("collection_name and vector are required")
	}
	return success("Vector inserted into " + collection)
}