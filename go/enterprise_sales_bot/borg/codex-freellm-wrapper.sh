#!/bin/bash
# Codex wrapper that uses FreeLLM endpoint
# This script wraps codex CLI to use the local FreeLLM server

export OPENAI_BASE_URL=http://localhost:4000/v1
export OPENAI_API_KEY=sk-freellm

# Default model that works with FreeLLM
DEFAULT_MODEL="poolside/laguna-xs.2:free"

# Build the command
CMD="codex"
if [ "$1" = "exec" ]; then
    # Check if --model is already specified
    if [[ "$*" == *"--model"* ]]; then
        CMD="codex $*"
    else
        # Insert --model after exec
        CMD="codex exec --model \"$DEFAULT_MODEL\" \"${@:2}\""
    fi
else
    CMD="codex --model \"$DEFAULT_MODEL\" \"$*\""
fi

echo "Executing: $CMD"
eval "$CMD"
