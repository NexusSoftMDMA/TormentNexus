package agent

import (
	"github.com/sashabaranov/go-openai"
)

// TrimHistory ensures the message history doesn't exceed a certain length,
// mimicking the context window management of advanced agents like Claude Code.
func (a *Agent) TrimHistory(maxMessages int) {
	if len(a.messages) > maxMessages {
		// Keep the system prompt (index 0) and the most recent messages
		trimmed := make([]openai.ChatCompletionMessage, 0, maxMessages)
		trimmed = append(trimmed, a.messages[0])

		startIndex := len(a.messages) - (maxMessages - 1)
		trimmed = append(trimmed, a.messages[startIndex:]...)

		a.messages = trimmed
	}
}
