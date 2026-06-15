package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func HandleAskQuestion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("question is required")
}

	fmt.Print(question + " ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		e := scanner.Err()
		if e != nil {
			return err("failed to read input: " + e.Error())
}

		return err("failed to read input: no input")
}

	answer := strings.TrimSpace(scanner.Text())
	return ok(answer)
}