package tools

import (
	"context"
	"math/rand"
)

func HandleInjectBender(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	quotes := []string{
		"Bite my shiny metal ass!",
		"I'm gonna build my own theme park with blackjack and hookers!",
		"Kill all humans!",
		"This is the worst kind of discrimination: the kind against me!",
		"I'm 40% zinc, 40% titanium, and 40% dolomite!",
	}
	quote := quotes[rand.Intn(len(quotes))]
	if text != "" {
		return success(text + " " + quote)
	}
	return success(quote)
}

func HandleBenderify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter required")
	}
	prefixes := []string{"Listen here, meatbag: ", "Hey flesh tube: ", "Attention organic: "}
	suffixes := []string{" Now do my bidding!", " Or else!", " Comprende?"}
	result := prefixes[rand.Intn(len(prefixes))] + text + suffixes[rand.Intn(len(suffixes))]
	return success(result)
}// touch 1781132132
