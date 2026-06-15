package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetMeetingNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing meeting id")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/notes/"+id, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("notes retrieved")
}

func HandleSearchCalendar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/calendar?q="+query, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	return success("calendar search complete")
}// touch 1781132127
