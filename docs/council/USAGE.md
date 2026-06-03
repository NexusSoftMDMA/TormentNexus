# Usage Examples

This document provides practical examples of using the cli-orchestrator council plugin.

## Basic Setup

1. **Install the plugin**:
   ```bash
   pnpm install
   ```

   The council now ships inside the tormentnexus monorepo as the **cli-orchestrator** lane rather than as a standalone `tormentnexus-orchestrator-council` package.

2. **Create configuration**:
   Create `.opencode/council.json` in your project:
   ```json
   {
     "supervisors": [
       {
         "name": "ChatGPT",
         "provider": "openai",
         "model": "gpt-4"
       },
       {
         "name": "Claude",
         "provider": "anthropic",
         "model": "claude-3-5-sonnet-20241022"
       }
     ],
     "debateRounds": 2,
     "consensusThreshold": 0.5
   }
   ```

3. **Set environment variables**:
   ```bash
   export OPENAI_API_KEY="sk-..."
   export ANTHROPIC_API_KEY="sk-ant-..."
   ```

## Scenario 1: Automatic Code Review

The council automatically reviews code changes:

```typescript
// When you edit a file, the council will automatically review it
// Example: editing src/auth.ts

// The council will:
// 1. Detect the file edit
// 2. Each supervisor reviews the changes
// 3. Supervisors debate through multiple rounds
// 4. Final votes are cast
// 5. Consensus is reached
```

**Expected Output:**
```
🔍 Council reviewing file edit: src/auth.ts

🗳️  Council Debate - Round 1: Initial Opinions
📢 ChatGPT (openai) is evaluating...
✓ ChatGPT: The authentication implementation follows OAuth2 best practices...

📢 Claude (anthropic) is evaluating...
✓ Claude: Security measures are adequate, but I suggest adding rate limiting...

🗳️  Council Debate - Round 2: Deliberation
...

📊 Voting Results: 2/2 approved (100%)
Decision: ✅ APPROVED
```

## Scenario 2: Manual Debate

Trigger a manual debate on any topic:

```bash
# Using OpenCode CLI
opencode council_debate topic="Should we migrate from REST to GraphQL?"
```

**Response:**
```json
{
  "decision": "APPROVED",
  "consensus": "67%",
  "reasoning": "After 2 supervisor votes, the council has reached consensus...",
  "votes": [
    {
      "supervisor": "ChatGPT",
      "approved": true,
      "comment": "GraphQL provides better flexibility for our use case..."
    },
    {
      "supervisor": "Claude",
      "approved": true,
      "comment": "The migration would improve our API efficiency..."
    }
  ]
}
```

## Scenario 3: Check Council Status

```bash
opencode council_status
```

**Response:**
```json
{
  "enabled": true,
  "totalSupervisors": 2,
  "supervisors": [
    {
      "name": "ChatGPT",
      "provider": "openai"
    },
    {
      "name": "Claude",
      "provider": "anthropic"
    }
  ]
}
```

## Scenario 4: Multi-Provider Council

Use supervisors from different providers:

```json
{
  "supervisors": [
    {
      "name": "GPT-4",
      "provider": "openai",
      "model": "gpt-4",
      "systemPrompt": "Focus on performance and scalability"
    },
    {
      "name": "Claude",
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022",
      "systemPrompt": "Focus on security and best practices"
    },
    {
      "name": "DeepSeek",
      "provider": "deepseek",
      "model": "deepseek-chat",
      "baseURL": "https://api.deepseek.com",
      "systemPrompt": "Focus on code quality and maintainability"
    },
    {
      "name": "Grok",
      "provider": "grok",
      "model": "grok-beta",
      "baseURL": "https://api.x.ai/v1",
      "systemPrompt": "Focus on innovation and creativity"
    }
  ],
  "debateRounds": 3,
  "consensusThreshold": 0.75
}
```

## Scenario 5: Specialized Supervisors

Create supervisors with specialized roles:

```json
{
  "supervisors": [
    {
      "name": "Security Expert",
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022",
      "systemPrompt": "You are a security expert. Focus exclusively on finding security vulnerabilities, authentication issues, and potential exploits."
    },
    {
      "name": "Performance Guru",
      "provider": "openai",
      "model": "gpt-4",
      "systemPrompt": "You are a performance optimization specialist. Focus on code efficiency, algorithmic complexity, and resource usage."
    },
    {
      "name": "Architecture Reviewer",
      "provider": "deepseek",
      "model": "deepseek-chat",
      "baseURL": "https://api.deepseek.com",
      "systemPrompt": "You are a software architect. Focus on design patterns, SOLID principles, and long-term maintainability."
    }
  ],
  "debateRounds": 2,
  "consensusThreshold": 0.67
}
```

## Scenario 6: Toggle Council Monitoring

Temporarily disable the council:

```bash
# Disable monitoring
opencode council_toggle enabled=false

# Make changes without council review
# ...

# Re-enable monitoring
opencode council_toggle enabled=true
```

## Scenario 7: High-Stakes Changes

For critical changes, require unanimous approval:

```json
{
  "supervisors": [
    {
      "name": "Senior Engineer",
      "provider": "openai",
      "model": "gpt-4"
    },
    {
      "name": "Security Officer",
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022"
    },
    {
      "name": "Tech Lead",
      "provider": "deepseek",
      "model": "deepseek-chat",
      "baseURL": "https://api.deepseek.com"
    }
  ],
  "debateRounds": 3,
  "consensusThreshold": 1.0
}
```

## Scenario 8: Quick Reviews

For rapid development, use a single supervisor:

```json
{
  "supervisors": [
    {
      "name": "Code Reviewer",
      "provider": "openai",
      "model": "gpt-3.5-turbo"
    }
  ],
  "debateRounds": 1,
  "autoApprove": true,
  "consensusThreshold": 1.0
}
```

## Integration with CI/CD

You can integrate the council into your CI/CD pipeline:

```yaml
# .github/workflows/council-review.yml
name: Council Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
      
      - name: Install dependencies
        run: npm install
      
      - name: Run council review
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          opencode council_debate topic="Review PR changes"
```

## Best Practices

1. **Start Small**: Begin with 2-3 supervisors and increase as needed
2. **Specialized Roles**: Assign specific focus areas to each supervisor
3. **Adjust Thresholds**: Set `consensusThreshold` based on project criticality
4. **Custom Prompts**: Tailor system prompts to your team's standards
5. **Monitor Costs**: Be mindful of API costs with multiple supervisors
6. **Toggle When Needed**: Disable council for experimental branches
7. **Debate Rounds**: More rounds = more thorough but slower reviews

## Troubleshooting

### No supervisors available
- Check that API keys are set in environment variables
- Verify the config file path and format
- Check network connectivity to API endpoints

### Council not triggering
- Ensure the plugin is properly installed
- Check that hooks are registered correctly
- Verify the OpenCode plugin system is active

### Slow performance
- Reduce number of supervisors
- Decrease debate rounds
- Use faster models (e.g., gpt-3.5-turbo)

## Advanced Configuration

### Custom Webhooks

You can extend the council to send notifications:

```typescript
// In your custom plugin extension
council.on('decision', (decision) => {
  if (!decision.approved) {
    // Send notification to Slack, Discord, etc.
    notifyTeam(decision);
  }
});
```

### Persistent History

Store debate history for analysis:

```json
{
  "supervisors": [...],
  "debateRounds": 2,
  "consensusThreshold": 0.5,
  "historyPath": ".opencode/council-history.json"
}
```

## Support

For issues, questions, or contributions:
- GitHub Issues: use the main tormentnexus repository issue tracker.
- Documentation: [README.md](../README.md)
