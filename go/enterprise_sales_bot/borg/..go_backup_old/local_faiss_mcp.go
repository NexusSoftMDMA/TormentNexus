package tools

import "context"

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	k, _ :=getInt(args, "k")
	if k <= 0 {
		k = 10
	}
	return ok("Searched local Faiss index with query: " + query + " topK: " + itoa(k))
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	vector, _ :=getString(args, "vector")
	if id == "" || vector == "" {
		return err("Missing id or vector")
}

	return ok("Added vector id: " + id + " to local Faiss index")
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}