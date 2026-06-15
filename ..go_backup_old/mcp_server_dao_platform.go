package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCreateProposal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	description, _ :=getString(args, "description")
	body := map[string]string{"title": title, "description": description}
	data, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal proposal")
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dao.com/proposals", bytes.NewReader(data))
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("proposal creation failed")
	}
	return success("proposal created")
}

func HandleVote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	proposalID, _ :=getString(args, "proposalId")
	support, _ :=getBool(args, "support")
	body := map[string]interface{}{"proposalId": proposalID, "support": support}
	data, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal vote")
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dao.com/votes", bytes.NewReader(data))
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("vote failed")
	}
	io.Copy(io.Discard, resp.Body)
	return success("vote recorded")
}