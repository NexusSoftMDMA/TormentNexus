#!/bin/bash
# Codex wrapper that defaults to FreeLLM endpoint
# This script ensures codex always uses the FreeLLM router

# Set FreeLLM environment variables
export OPENAI_BASE_URL=http://localhost:4000/v1
export OPENAI_API_KEY=sk-freellm

# Default model (can be overridden with --model flag)
DEFAULT_MODEL="poolside/laguna-xs.2:free"

# Build arguments
ARGS=("$@")
HAS_MODEL=false

for arg in "${ARGS[@]}"; do
    if [[ "$arg" == "--model" ]] || [[ "$arg" == "-m" ]]; then
        HAS_MODEL=true
    fi
done

# If no model specified, add default
if [ "$HAS_MODEL" = false ]; then
    ARGS+=(--model "$DEFAULT_MODEL")
fi

# Execute codex with the configured environment
exec codex "${ARGS[@]}"
