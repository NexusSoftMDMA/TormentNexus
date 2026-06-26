# Golden Tool Contracts — Built-In Tools Evidence Fixtures

> Purpose: Canonical I/O contracts for built-in tools across platforms.  
> Last updated: 2026-03-19  
> Status: Fixture scaffolds (L2 grounding)

## Overview

This file maps exact tool call signatures, argument schemas, and return payloads across platforms captured in [BUILTIN_TOOLS_EVIDENCE_LOCK.md](BUILTIN_TOOLS_EVIDENCE_LOCK.md). Each section represents a **tool contract** that serves as:

1. **Equivalence baseline** for cross-platform parity testing
2. **CI validation gate** to catch contract drift on release updates
3. **Golden reference** for agent tool-use standardization

---

## Platform: GitHub Copilot CLI (L2)

### Tool: `bash` / `powershell`

**Signature:**
```
bash(command: string, timeout?: number) -> { exitCode: number; stdout: string; stderr: string }
powershell(command: string, timeout?: number) -> { exitCode: number; stdout: string; stderr: string }
```

**Permission pattern:**
```
--allow-tool='shell(git:*)'     # Allow all git subcommands
--allow-tool='shell(npm test)'  # Allow exact command
--deny-tool='shell(rm -rf *)'   # Block destructive
```

**Hook contract (preToolUse):**
```json
{
  "permissionDecision": "allow" | "deny" | "ask",
  "permissionDecisionReason": "string",
  "modifiedArgs": { "command": "string" }
}
```

**Evidence source:**
- https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference
- Hook reference: `preToolUse` decision control

---

### Tool: `view` (read file)

**Signature:**
```
view(path: string) -> { content: string; encoding: string }
```

**Permission pattern:**
```
--allow-tool='read'              # Allow all reads
--allow-tool='read(.env)'        # Allow specific file
--deny-tool='read(.git/**)'      # Block reads under .git
```

**Evidence source:**
- CLI command reference, "Tool names for hook matching"

---

### Tool: `edit` / `create` (write file)

**Signature:**
```
edit(path: string, content: string) -> { success: boolean; path: string; length: number }
create(path: string, content: string) -> { success: boolean; path: string; length: number }
```

**Permission pattern:**
```
--allow-tool='write'                    # Allow all writes
--allow-tool='write(.github/**)'        # Allow writes to .github
--deny-tool='write(.git/**)'            # Never allow .git writes
```

**Evidence source:**
- CLI command reference, "Tool permission patterns"

---

## Platform: Codex CLI (L2)

### Tool: `bash` (shell execution)

**Signature:**
```
bash(command: string, options?: { cwd?: string; env?: Record<string, string>; timeout?: number }) 
  -> { exitCode: number; stdout: string; stderr: string; duration_ms: number }
```

**Sandbox modes:**
```
--sandbox read-only              # No shell execution (read-only)
--sandbox workspace-write        # Write to workspace, no network
--sandbox danger-full-access     # Full permissions (unsafe)
--full-auto                      # Shorthand: workspace-write + on-request approvals
```

**Approval policy:**
```
--ask-for-approval on-request    # Ask before risky ops (default for workspace-write)
--ask-for-approval never         # Auto-approve all (CI/headless)
--ask-for-approval untrusted     # Ask for mutation/network only
```

**Evidence source:**
- https://developers.openai.com/codex/cli/reference
- https://developers.openai.com/codex/agent-approvals-security

---

### Tool: `/permissions` (dynamic approval control)

**Signature:**
```
/permissions
  -> interactive UI to select preset (Auto | Read-Only | Full-Access)
  -> updates session approval policy in real-time
```

**Return effect:**
```
session.approvalPolicy = selected_policy
session.sandbox = mapped_sandbox_for_policy
```

**Evidence source:**
- https://developers.openai.com/codex/cli/slash-commands

---

## Platform: Claude Code (L2)

### Hook: `PreToolUse`

**Signature:**
```
PreToolUse(tool_name: string, tool_input: object) 
  -> { 
    permissionDecision: "allow" | "deny" | "ask",
    permissionDecisionReason?: string,
    updatedInput?: object,
    additionalContext?: string
  }
```

**Tool matcher values:**
```
Bash              # Shell execution
Edit | Write      # File modification
Read | Glob | Grep # File reads
WebFetch          # Web access
Agent(AgentName)  # Subagent invocation
mcp__*__*         # MCP tool pattern
```

**Exit code semantics:**
```
0                 # Allow (process JSON output for fine-grained control)
2                 # Deny blocking error (tool call prevented)
other non-zero    # Log and continue (non-blocking)
```

**Evidence source:**
- https://code.claude.com/docs/en/hooks
- Hook lifetime and matcher/decision control sections

---

### Permission rule syntax

**Signature:**
```
Rule = Tool(specifier)? where specifier filters on:
  - Bash commands: Bash(npm run *), Bash(git commit *)
  - File paths: Read(*.env), Edit(/src/**)
  - Domains: WebFetch(domain:github.com)
  - MCP servers: mcp__server__tool_name
```

**Precedence:**
```
deny > ask > allow  // First matching rule wins
```

**Evidence source:**
- https://code.claude.com/docs/en/permissions

---

## Platform: Cursor (L2)

### Hook: `preToolUse`

**Signature:**
```
preToolUse(hook_event_name: "preToolUse", tool_name: string, tool_input: object)
  -> { 
    permissionDecision: "allow" | "deny" | "ask",
    permissionDecisionReason?: string
  }
```

**Tool category matchers:**
```
Shell      # Bash/shell execution
Read       # File reads
Write      # File creation/edit
Grep       # Search
Delete     # File deletion
Task       # Background tasks
MCP:*      # MCP tool invocation
```

**Config schema (hooks.json):**
```json
{
  "hooks": {
    "preToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "path/to/script.sh",
            "timeout": 600000
          }
        ]
      }
    ]
  }
}
```

**Evidence source:**
- Cursor hooks reference (captured in current evidence-gather pass)

---

## Validation Checklist

- [ ] Each tool contract includes signature with types
- [ ] Permission/approval pattern documented per platform
- [ ] Hook decision fields and exit-code semantics normalized
- [ ] Evidence sources (first-party URL + date) pinned
- [ ] Fixture used in CI to catch contract drift on release updates
- [ ] Cross-platform equivalence matrix created (e.g., bash→bash, edit→Edit/Write, etc.)

---

## Next Steps (L3 Promotion Path)

1. **Add response payload examples** per tool per platform
2. **Create CI test fixtures** that validate contract compliance
3. **Document tool-use equivalence** (e.g., semantic mapping between Codex `/permissions` and Claude Code permission modes)
4. **Release-pin verification** (capture version output on each platform and validate against locked contract)

---

## Related Files

- [BUILTIN_TOOLS_EVIDENCE_LOCK.md](BUILTIN_TOOLS_EVIDENCE_LOCK.md) — Platform lock state and evidence sources
- [DESIGN_STANDARDS.md](../DESIGN_STANDARDS.md) — Agent design and tool-use best practices
