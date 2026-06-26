@echo off
REM Codex wrapper for FreeLLM endpoint on Windows
REM This script ensures codex always uses the FreeLLM router

set OPENAI_BASE_URL=http://localhost:4000/v1
set OPENAI_API_KEY=sk-freellm

REM Execute codex with FreeLLM configuration
REM Note: Not specifying --model lets codex use its default routing
codex %*

REM To force a specific model, uncomment and modify:
REM codex --model claude-3-7-sonnet-20250219 %*