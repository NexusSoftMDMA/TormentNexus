package tools

var store = make(map[string]map[string][]string)

func HandleAddTriple(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subj, _ :=getString(args, "subject")
	pred, _ :=getString(args, "predicate")
	obj, _ :=getString(args, "object")
	if subj == "" || pred == "" || obj == "" {
		return err("missing required fields")
}

	if store[subj] == nil {
		store[subj] = make(map[string][]string)

	store[subj][pred] = append(store[subj][pred], obj)
	return ok("triple added")
}

}

func HandleGetTriples(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subj, _ :=getString(args, "subject")
	pred, _ :=getString(args, "predicate")
	if subj == "" {
		return err("subject required")
}

	sub, found := store[subj]
	if !found {
		return err("subject not found")
}

	if pred == "" {
		return success(sub)
}

	objs, found := sub[pred]
	if !found {
		return err("predicate not found")
}

	return success(objs)
}