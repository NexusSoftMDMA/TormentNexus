package mcpimpl

import (
	"context"
	"fmt"
)

func HandleIsPrime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	n, _ :=getInt(args, "number")
	if n < 2 {
		return err("number must be >= 2")
}

	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return ok(fmt.Sprintf("%d is not prime", n))

	}
	return ok(fmt.Sprintf("%d is prime", n))
}

}

func HandleNextPrime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	n, _ :=getInt(args, "number")
	candidate := n + 1
	for {
		if candidate < 2 {
			candidate = 2
		}
		isPrime := true
		for i := 2; i*i <= candidate; i++ {
			if candidate%i == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			return ok(fmt.Sprintf("Next prime after %d is %d", n, candidate))
}

		candidate++
	}
}