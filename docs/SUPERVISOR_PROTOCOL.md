# Supervisor Nudge Protocol

## Purpose
The **Supervisor Nudge Protocol** is an autonomous mechanism designed to maintain development momentum by addressing AI agent inactivity. When an agent becomes stalled or silent beyond defined thresholds, the system provides a short, professional directive to re-engage the agent.

## Core Logic
The protocol follows these constraints to ensure effective re-engagement without context bloat:
1. **Inactivity Detection**: The system monitors session `updated_at` timestamps against `InactivityThresholdMinutes` (default) and `ActiveWorkThresholdMinutes` (for IN_PROGRESS tasks).
2. **Short & Direct**: Nudges are kept brief and professional, acting as a high-level user instruction rather than a meta-commentary on the agent's state.
3. **Professional Persona**: The supervisor speaks as if they are the user giving a direct command, avoiding self-referential language about being a "supervisor."
4. **Context-Aware**: The nudge is generated based on the recent conversation context to provide a relevant "next step" or encouragement.

## Prompting Style
Nudges are generated using a specific system prompt:
> "You are a supervisor for an AI agent. The agent has been inactive. Your goal is to provide a helpful, encouraging nudge or a specific instruction based on the recent context to get the agent moving again. Keep the message short, direct, and professional. Do not mention that you are a supervisor. Just speak as if you are the user giving a command."

## Implementation Reference
- **TypeScript**: `packages/ui/src/lib/orchestration/supervisor.ts`
- **Go Orchestrator**: `orchestrator/queue_workers.go` (Handles inactivity detection and trigger logic).

## Integration
Nudges are logged as `action` types in the `KeeperLog` table, ensuring a clear audit trail of autonomous re-engagement events.
