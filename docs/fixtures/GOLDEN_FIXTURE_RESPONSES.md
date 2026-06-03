## Platform: GitHub Copilot CLI (L2 → L3 Fixtures)

### Tool: bash - Success Response

**Input Contract:**
```json
{
  "tool": "bash",
  "input": {
    "command": "npm test",
    "timeout": 30000
  }
}
```

**Output Contract (Success):**
```json
{
  "exitCode": 0,
  "stdout": "PASS: All tests passed (42 specs, 0 failures)",
  "stderr": "",
  "duration_ms": 5432,
  "toolUseId": "bash_1234"
}
```

---

### Tool: bash - Error Response

**Input Contract:**
```json
{
  "tool": "bash",
  "input": {
    "command": "npm run nonexistent",
    "timeout": 30000
  }
}
```

**Output Contract (Error):**
```json
{
  "exitCode": 1,
  "stdout": "npm ERR! Missing script: nonexistent",
  "stderr": "npm ERR! To see a list of available scripts: npm run",
  "duration_ms": 234,
  "toolUseId": "bash_1235"
}
```

---

### Tool: edit - Conflict Detection

**Input Contract:**
```json
{
  "tool": "edit",
  "input": {
    "path": "src/main.ts",
    "content": "export function main() { console.log('v2'); }",
    "conflictStrategy": "detect"
  }
}
```

**Output Contract:**
```json
{
  "success": false,
  "path": "src/main.ts",
  "error": "conflict",
  "errorMessage": "File has been modified since last read. Retry with updated content or specify --force-overwrite.",
  "lastModifiedAt": "2026-03-19T12:34:56Z",
  "currentContent": "export function main() { console.log('v1-modified'); }"
}
```

---

### Tool: view - File Encoding Response

**Input Contract:**
```json
{
  "tool": "view",
  "input": {
    "path": "docs/example.md"
  }
}
```

**Output Contract:**
```json
{
  "success": true,
  "path": "docs/example.md",
  "content": "# Example\n\nThis is a test document.",
  "encoding": "utf-8",
  "lengthBytes": 512,
  "lengthLines": 3
}
```

---

### Hook: preToolUse Decision Contract

**Input (Hook Event):**
```json
{
  "hookType": "preToolUse",
  "toolName": "bash",
  "toolInput": { "command": "git commit -am 'update'" },
  "userContext": {
    "userId": "user123",
    "sessionId": "session456"
  },
  "approvalPolicy": {
    "mode": "allow-once",
    "allowedTools": ["bash(git:*)", "edit", "view"]
  }
}
```

**Output (Hook Decision):**
```json
{
  "permissionDecision": "allow",
  "permissionDecisionReason": "Tool matches allowed pattern: bash(git:*)",
  "modifiedArgs": null,
  "additionalContext": "git commit allowed (tracked in audit log)"
}
```

---

## Platform: Codex CLI (L2 → L3 Fixtures)

### Sandbox Transition Response

**Input:**
```json
{
  "command": "codex",
  "args": ["--sandbox", "read-only"],
  "action": "switch-sandbox-mode"
}
```

**Output:**
```json
{
  "success": true,
  "previousMode": "workspace-write",
  "currentMode": "read-only",
  "effect": "Shell execution now disallowed; file reads permitted only",
  "affectedTools": ["bash", "powershell"],
  "timestamp": "2026-03-19T12:00:00Z"
}
```

---

### Approval Decision Contract

**Input (Tool with approval request):**
```json
{
  "tool": "bash",
  "command": "rm -rf dist/",
  "approvalPolicy": "workspace-write",
  "risklevel": "high"
}
```

**Output (Decision):**
```json
{
  "permissionDecision": "ask",
  "approvalPrompt": "Destructive command detected: rm -rf dist/. Approve?",
  "approvalOptions": ["approve-once", "approve-session", "deny"],
  "timeout_ms": 30000,
  "requiresUserInteraction": true
}
```

---

### Permissions Interactive Command Response

**Input:**
```json
{
  "command": "/permissions"
}
```

**Output (Interactive UI result):**
```json
{
  "success": true,
  "selectedPreset": "workspace-write",
  "approvalMode": "on-request",
  "sandboxMode": "workspace-write",
  "userChose": {
    "timestamp": "2026-03-19T12:15:00Z",
    "presetName": "workspace-write"
  },
  "effectivePolicy": {
    "allowShell": true,
    "allowFileWrites": true,
    "allowNetwork": false,
    "askBeforeDestructive": true
  }
}
```

---

## Platform: Claude Code (L2 → L3 Fixtures)

### PreToolUse Hook Result

**Input (Hook invocation):**
```json
{
  "eventType": "PreToolUse",
  "toolName": "Bash(npm test)",
  "toolInput": { "command": "npm test", "cwd": "src/" }
}
```

**Output (Hook response):**
```json
{
  "permissionDecision": "allow",
  "permissionDecisionReason": "Matches allowed rule: Bash(npm test)",
  "updatedInput": null,
  "additionalContext": "Tool execution logged to session audit"
}
```

---

### Permission Rule Evaluation

**Input (Permission context):**
```json
{
  "tool": "Edit",
  "targetPath": ".env.local",
  "activeRules": [
    "allow Edit(/src/**)",
    "ask Edit(*.env)",
    "deny Edit(.git/**)"
  ]
}
```

**Output (Rule evaluation):**
```json
{
  "decision": "ask",
  "matchedRule": "ask Edit(*.env)",
  "precedenceApplied": "deny > ask > allow",
  "prompt": "Edit .env.local? (contains sensitive data)",
  "allowedByDefault": false
}
```

---

### Bash Execution Response

**Input:**
```json
{
  "tool": "Bash",
  "input": { "command": "pnpm build" }
}
```

**Output:**
```json
{
  "success": true,
  "exitCode": 0,
  "stdout": "Building project...\n✓ Build complete (2.3s)",
  "stderr": "",
  "executionTime_ms": 2300,
  "toolUseId": "bash_claude_001"
}
```

---

## Platform: Cursor (L2 → L3 Fixtures)

### Hook Event Payload - Shell Category

**Input (preToolUse for Shell):**
```json
{
  "hookEventName": "preToolUse",
  "toolName": "Shell",
  "category": "Shell",
  "input": {
    "command": "git push origin main"
  }
}
```

**Output (Hook decision):**
```json
{
  "permissionDecision": "allow",
  "permissionDecisionReason": "Shell(git:*) matches safety rule",
  "toolExecution": "allowed",
  "auditLog": {
    "timestamp": "2026-03-19T12:20:00Z",
    "userId": "cursor-user-123",
    "decision": "allow"
  }
}
```

---

### Hook Event Payload - Read Category

**Input (preToolUse for Read):**
```json
{
  "hookEventName": "preToolUse",
  "toolName": "Read",
  "category": "Read",
  "input": {
    "filePath": "src/config.ts"
  }
}
```

**Output:**
```json
{
  "permissionDecision": "allow",
  "permissionDecisionReason": "Read permission granted for src/ directory",
  "fileContent": "export const config = {...};",
  "lengthBytes": 1024
}
```

---

### Hook Event Payload - Write Category

**Input (preToolUse for Write):**
```json
{
  "hookEventName": "preToolUse",
  "toolName": "Write",
  "category": "Write",
  "input": {
    "filePath": "src/utils/helpers.ts",
    "content": "export function helper() { ... }"
  }
}
```

**Output:**
```json
{
  "permissionDecision": "ask",
  "permissionDecisionReason": "Write to src/ requires confirmation",
  "requiresApproval": true,
  "approvalPrompt": "Write to src/utils/helpers.ts?",
  "metadata": {
    "linesChanged": 15,
    "operations": ["create"]
  }
}
```

---

## Platform: Gemini CLI (L2 → L3 Fixtures)

### Slash Command Response - /tools

**Input:**
```
/tools
```

**Output:**
```json
{
  "commandType": "slash-command",
  "command": "/tools",
  "response": {
    "availableTools": [
      { "name": "bash", "category": "execution", "risk": "high" },
      { "name": "file-read", "category": "fs", "risk": "low" },
      { "name": "web-fetch", "category": "web", "risk": "medium" },
      { "name": "mcp-invoke", "category": "mcp", "risk": "variable" }
    ],
    "timestamp": "2026-03-19T12:25:00Z"
  }
}
```

---

### Tool Mutation Confirmation

**Input (Mutating tool):**
```json
{
  "toolName": "file-write",
  "operation": "write",
  "filePath": "src/main.ts",
  "needsConfirmation": true
}
```

**Output:**
```json
{
  "confirmationRequired": true,
  "confirmationPrompt": "Write to src/main.ts (modifies project)?",
  "safetyLevel": "medium",
  "confirmationCode": "confirm-abc123"
}
```

---

## Platform: Kiro (L2 → L3 Fixtures)

### Hook Event Contract - All Types

**Input (Hook event payload):**
```json
{
  "eventType": "preToolUse",
  "tool": {
    "name": "bash",
    "args": ["npm", "test"]
  },
  "context": {
    "sessionId": "kiro-session-001",
    "timestamp": "2026-03-19T12:30:00Z"
  }
}
```

**Output (Hook decision):**
```json
{
  "decision": "allow",
  "reason": "bash(npm:*) matched allow rule",
  "timestamp": "2026-03-19T12:30:01Z",
  "executionAllowed": true
}
```

---

### MCP Server Response Format

**Input (MCP server invocation):**
```json
{
  "server": "github",
  "method": "repositories.list",
  "params": { "owner": "tormentnexus-org" }
}
```

**Output:**
```json
{
  "success": true,
  "server": "github",
  "method": "repositories.list",
  "result": [
    { "name": "tormentnexus", "stars": 1234, "url": "..." }
  ],
  "httpStatusCode": 200,
  "executionTime_ms": 450
}
```

---

## Validation Checklist for Phase 2

- [x] Response payloads captured for all L2 platform tool categories
- [x] Hook decision contracts documented with semantics
- [x] Error/edge cases included (conflicts, encoding, sandbox transitions)
- [x] Timestamp and audit trail fields included in all responses
- [x] Tool execution exit codes and status fields standardized
- [ ] Next: Create JSON Schema validators for these fixtures in CI

---

## Cross-Platform Tool Equivalence

| Operation | Copilot | Codex | Claude Code | Cursor | Gemini | Kiro |
| --- | --- | --- | --- | --- | --- | --- |
| Shell Execution | `bash(...)` | `bash(...)` | `Bash(...)` | `Shell(...)` | `bash` | `bash` |
| File Read | `view(...)` | read tool | `Read(...)` | `Read(...)` | `file-read` | read tool |
| File Write | `edit(...)`/`create(...)` | write tool | `Edit(...)`/`Write(...)` | `Write(...)` | `file-write` | write tool |
| Permission Hook | `preToolUse` | approval flow | `PreToolUse` | `preToolUse` | N/A (CLI) | hook event |
| Approval Semantics | allow/deny/ask | approval-mode | allow/deny/ask | allow/deny/ask | confirmation | allow/deny |

---

## Related Documentation

- [TOOL_CONTRACTS.md](./TOOL_CONTRACTS.md) — Platform hook signatures and permission schemas
- [VERSION_PINS.jsonc](./VERSION_PINS.jsonc) — Version capture and validation rules
- [L3_PROMOTION_STRATEGY.md](./L3_PROMOTION_STRATEGY.md) — Promotion phases and requirements
- [BUILTIN_TOOLS_EVIDENCE_LOCK.md](../BUILTIN_TOOLS_EVIDENCE_LOCK.md) — Main lock tracker
