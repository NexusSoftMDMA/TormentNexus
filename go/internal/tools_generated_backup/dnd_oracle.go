package tools

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())

}

func HandleRollDice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diceStr, _ :=getString(args, "dice")
	if diceStr == "" {
		return err("dice parameter required (e.g., '2d6')")
}

	parts := strings.Split(diceStr, "d")
	if len(parts) != 2 {
		return err("invalid dice format, use NdM")
}

	count, e1 := strconv.Atoi(parts[0])
	sides, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || count < 1 || sides < 1 {
		return err("invalid dice specification")
}

	total := 0
	rolls := make([]int, count)
	for i := 0; i < count; i++ {
		roll := rand.Intn(sides) + 1
		rolls[i] = roll
		total += roll
	}
	return ok(fmt.Sprintf("Rolled %s: %v (total %d)", diceStr, rolls, total))
}

func HandleOracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("question parameter required")
}

	answers := []string{"Yes", "No", "Maybe", "Ask again later", "Definitely", "Unlikely"}
	idx := rand.Intn(len(answers))
	return ok(fmt.Sprintf("Oracle says: %s", answers[idx]))
}