# OpenAI Codex with FreeLLM Integration

## Status: ✅ CONFIGURED

### Overview
This document describes the integration of OpenAI Codex CLI with the FreeLLM local endpoint running at `http://localhost:4000/v1`.

### Configuration

#### Environment Variables
```bash
export OPENAI_BASE_URL=http://localhost:4000/v1
export OPENAI_API_KEY=sk-freellm
```

#### Model Selection
The FreeLLM endpoint routes requests through OpenRouter's free tier. The model `poolside/laguna-xs.2:free` is available but requires proper routing.

### Usage

#### Direct Codex Command
```bash
# Set environment
export OPENAI_BASE_URL=http://localhost:4000/v1
export OPENAI_API_KEY=sk-freellm

# Run codex (will default to gpt-5.5 due to FreeLLM routing)
codex exec "your prompt here"
```

#### Shell Wrapper Script
A wrapper script `codex-freellm-wrapper.sh` is provided for convenience:
```bash
./codex-freellm-wrapper.sh exec "your prompt here"
```

### API Endpoint Verification

The FreeLLM endpoint is operational:
```
POST http://localhost:4000/v1/chat/completions
Headers:
  - Content-Type: application/json
  - Authorization: Bearer sk-freellm
```

### Available Models via FreeLLM
- `poolside/laguna-xs.2:free` (target model)
- Routes to: `moonshotai/kimi-k2.6`, `openrouter/free`, etc.

### Known Limitations
1. **Model Mapping**: FreeLLM may route to different models than specified
2. **Quota**: The endpoint may have usage limits
3. **Codex Default**: Codex CLI defaults to `gpt-5.5` when no model is specified

### Next Steps
1. Verify model routing in FreeLLM configuration
2. Test Codex with specific prompts
3. Monitor usage and quota

---
*🤖 Model: poolside/laguna-xs.2:free | 🔌 Provider: openrouter*
