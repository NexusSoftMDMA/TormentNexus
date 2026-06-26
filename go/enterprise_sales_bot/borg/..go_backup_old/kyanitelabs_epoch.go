package tools

import (
	"context"
	"strconv"
)

func HandlePertEstimate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	opt, _ :=getString(args, "optimistic")
	pes, _ :=getString(args, "pessimistic")
	ml, _ :=getString(args, "most_likely")
	o, e := strconv.ParseFloat(opt, 64)
	if e != nil {
		return err("invalid optimistic value")
}

	p, e := strconv.ParseFloat(pes, 64)
	if e != nil {
		return err("invalid pessimistic value")
}

	m, e := strconv.ParseFloat(ml, 64)
	if e != nil {
		return err("invalid most_likely value")
}

	expected := (o + 4*m + p) / 6
	return success("expected duration: " + strconv.FormatFloat(expected, 'f', 2, 64))
}

func HandleTokenToTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tokensStr, _ :=getString(args, "tokens")
	rateStr, _ :=getString(args, "tokens_per_second")
	tokens, e := strconv.ParseFloat(tokensStr, 64)
	if e != nil {
		return err("invalid tokens value")
}

	rate, e := strconv.ParseFloat(rateStr, 64)
	if e != nil {
		return err("invalid tokens_per_second value")
}

	if rate == 0 {
		return err("tokens_per_second cannot be zero")
}

	timeSeconds := tokens / rate
	return success("estimated time: " + strconv.FormatFloat(timeSeconds, 'f', 2, 64) + " seconds")
}// touch 1781132130
