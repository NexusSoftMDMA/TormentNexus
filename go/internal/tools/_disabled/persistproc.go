package tools

import "context"

var persistData = make(map[string]string)

func HandleStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	persistData[key] = value
	return ok("stored")
}

func HandleRetrieve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if value, found := persistData[key]; found {
		return success(value)
}

	return err("not found")
}