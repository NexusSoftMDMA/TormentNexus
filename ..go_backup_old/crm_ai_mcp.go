package tools

import (
	"context"
	"strconv"
)

func HandleLeadScorer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	company, _ :=getString(args, "company")
	email, _ :=getString(args, "email")
	score := 50
	if company == "AI" || company == "AI Labs" {
		score = 90
	}
	msg := "Lead " + name + " (" + email + ") from " + company + " scored " + strconv.Itoa(score)
	return success(msg)
}

func HandleDealStagePredictor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	stage, _ :=getString(args, "stage")
	var predicted string
	if stage == "Qualification" && amount > 10000 {
		predicted = "Proposal"
	} else if stage == "Proposal" && amount > 50000 {
		predicted = "Negotiation"
	} else {
		predicted = "Closed Won"
	}
	msg := "Deal of $" + strconv.Itoa(amount) + " at stage '" + stage + "' predicted next: " + predicted
	return ok(msg)
}