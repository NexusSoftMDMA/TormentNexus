package ctxharvester

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ContextGroomer struct {
	maxTokens int
}

func NewContextGroomer(maxTokens int) *ContextGroomer {
	if maxTokens <= 0 {
		maxTokens = 8000
	}
	return &ContextGroomer{maxTokens: maxTokens}
}

func (g *ContextGroomer) CompressContext(messages []ChatMessage) []ChatMessage {
	// Very rudimentary token estimation (approx 4 chars per token)
	estimateTokens := func(text string) int {
		return (len(text) + 3) / 4
	}

	currentTokens := 0
	var result []ChatMessage

	// Always keep the system message
	var systemMessage *ChatMessage
	for _, m := range messages {
		if m.Role == "system" {
			systemMessage = &m
			break
		}
	}

	if systemMessage != nil {
		currentTokens += estimateTokens(systemMessage.Content)
		result = append(result, *systemMessage)
	}

	// Filter out system messages for the recent kept list
	var nonSystem []ChatMessage
	for _, m := range messages {
		if m.Role != "system" {
			nonSystem = append(nonSystem, m)
		}
	}

	// Keep the most recent messages that fit
	var keptRecent []ChatMessage
	for i := len(nonSystem) - 1; i >= 0; i-- {
		msg := nonSystem[i]
		tokens := estimateTokens(msg.Content)
		if currentTokens+tokens > g.maxTokens {
			// Insert a summary placeholder if we run out of room
			keptRecent = append([]ChatMessage{{
				Role:    "system",
				Content: "[System Note: Earlier conversation context was compressed/pruned due to length limits.]",
			}}, keptRecent...)
			break
		}
		currentTokens += tokens
		keptRecent = append([]ChatMessage{msg}, keptRecent...)
	}

	return append(result, keptRecent...)
}
