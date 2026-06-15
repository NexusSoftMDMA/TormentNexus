package tools

import "context"

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agents := map[string]string{
		"agent1": "LangGraph Next.js Agent",
		"agent2": "Demo Assistant",
	}
	return ok(agents)
}

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	convoID, _ :=getString(args, "conversation_id")
	reply := "Agent reply to: " + message
	if convoID != "" {
		reply += " (conversation: " + convoID + ")"
	}
	return success(reply)
}