package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetCourses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.vorim.example/courses")
	if e != nil {
		return err("failed to fetch courses")
}

	defer resp.Body.Close()
	var courses []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&courses); e != nil {
		return err("failed to decode courses")
}

	return ok("fetched courses")
}

func HandleGetCourse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	resp, e := http.DefaultClient.Get("https://api.vorim.example/courses/" + id)
	if e != nil {
		return err("failed to fetch course")
}

	defer resp.Body.Close()
	var course map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&course); e != nil {
		return err("failed to decode course")
}

	return ok("fetched course")
}