package tools

import (
	"context"
	"encoding/json"
)

type Bill struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type Member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Party string `json:"party"`
}

func HandleGetBill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	billID, _ :=getString(args, "billId")
	if billID == "" {
		return err("billId is required")
	}
	bill := Bill{ID: billID, Title: "Sample Bill", Status: "Introduced"}
	data, e := json.Marshal(bill)
	if e != nil {
		return err("failed to marshal bill")
	}
	return success(string(data))
}

func HandleGetMember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memberID, _ :=getString(args, "memberId")
	if memberID == "" {
		return err("memberId is required")
	}
	member := Member{ID: memberID, Name: "John Doe", Party: "Independent"}
	data, e := json.Marshal(member)
	if e != nil {
		return err("failed to marshal member")
	}
	return success(string(data))
}