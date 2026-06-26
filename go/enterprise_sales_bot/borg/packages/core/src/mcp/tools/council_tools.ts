
import { z } from "zod";
import { CouncilService } from "../../services/CouncilService.js";

export const createCouncilTools = (councilService: CouncilService) => [
    {
        name: "council_start_session",
        description: "Start a new debate session for The Council to discuss a complex topic.",
        inputSchema: z.object({
            topic: z.string().describe("The topic or question to be debated"),
        }),
        handler: async (args: { topic: string }) => {
            const session = councilService.startSession(args.topic);
            return {
                content: [{ type: "text", text: `Council Session started with ID: ${session.id}` }]
            };
        }
    },
    {
        name: "council_list_sessions",
        description: "List active Council debate sessions.",
        inputSchema: z.object({}),
        handler: async () => {
            const sessions = councilService
                .listSessions()
                .filter((s) => s.status === 'active');
            return {
                content: [{ type: "text", text: JSON.stringify(sessions, null, 2) }]
            };
        }
    },
    {
        name: "council_submit_opinion",
        description: "Submit an opinion or argument to an active Council session.",
        inputSchema: z.object({
            sessionId: z.string().describe("The ID of the council session"),
            agentId: z.string().describe("The ID or Persona of the agent submitting the opinion"),
            content: z.string().describe("The argument or opinion text")
        }),
        handler: async (args: { sessionId: string; agentId: string; content: string }) => {
            councilService.submitOpinion(args.sessionId, args.agentId, args.content);
            return {
                content: [{ type: "text", text: `Opinion submitted to session ${args.sessionId}` }]
            };
        }
    },
    {
        name: "council_read_session",
        description: "Read the full history of a Council session including opinions and votes.",
        inputSchema: z.object({
            sessionId: z.string().describe("The ID of the council session"),
        }),
        handler: async (args: { sessionId: string }) => {
            const session = councilService.getSession(args.sessionId);
            if (!session) {
                return { isError: true, content: [{ type: "text", text: "Session not found" }] };
            }
            return {
                content: [{ type: "text", text: JSON.stringify(session, null, 2) }]
            };
        }
    },
    {
        name: "council_advance_round",
        description: "Advance the debate to the next round.",
        inputSchema: z.object({
            sessionId: z.string(),
        }),
        handler: async (args: { sessionId: string }) => {
            councilService.advanceRound(args.sessionId);
            return {
                content: [{ type: "text", text: `Session ${args.sessionId} advanced to next round` }]
            };
        }
    },
    {
        name: "council_vote",
        description: "Cast a vote in the Council session.",
        inputSchema: z.object({
            sessionId: z.string(),
            agentId: z.string(),
            choice: z.string(),
            reason: z.string()
        }),
        handler: async (args: { sessionId: string; agentId: string; choice: string; reason: string }) => {
            councilService.castVote(args.sessionId, args.agentId, args.choice, args.reason);
            return {
                content: [{ type: "text", text: `Vote caused by ${args.agentId}` }]
            };
        }
    },
    {
        name: "council_conclude",
        description: "End a council session.",
        inputSchema: z.object({
            sessionId: z.string(),
        }),
        handler: async (args: { sessionId: string }) => {
            councilService.concludeSession(args.sessionId);
            return {
                content: [{ type: "text", text: `Session ${args.sessionId} concluded` }]
            };
        }
    }
];
