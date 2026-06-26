package mcpimpl

import "context"

func HandleListAgents_fullstack_langgraph_nextjs_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agents := map[string]string{
		"agent1": "LangGraph Next.js Agent",
		"agent2": "Demo Assistant",
	}
	return ok(agents)
}

func HandleChat_fullstack_langgraph_nextjs_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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