# OpenAI Codex with FreeLLM Integration (Windows)

## Status: ✅ CONFIGURED

### Quick Start

```cmd
cd C:\Users\hyper\workspace\tormentnexus
codex_freellm.bat exec "your prompt here"
```

### Environment

- **Endpoint**: `http://localhost:4000/v1`
- **API Key**: `sk-freellm`
- **Model**: `poolside/laguna-xs.2:free`

### Wrapper Script

The `codex_freellm.bat` script automatically sets the required environment variables and passes all arguments to codex.

### Manual Usage

```cmd
set OPENAI_BASE_URL=http://localhost:4000/v1
set OPENAI_API_KEY=sk-freellm
codex --model poolside/laguna-xs.2:free [your arguments]
```