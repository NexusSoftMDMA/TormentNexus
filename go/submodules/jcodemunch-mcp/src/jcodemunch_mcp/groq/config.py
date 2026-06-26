"""Configuration for the gcm CLI — API keys, model defaults, env vars."""

import os
from dataclasses import dataclass, field
from typing import Optional


# Default models ranked by quality/speed trade-off
DEFAULT_MODEL = "llama-3.3-70b-versatile"
FAST_MODEL = "llama-3.1-8b-instant"

GROQ_BASE_URL = "https://api.groq.com/openai/v1"

# Context budget defaults
DEFAULT_TOKEN_BUDGET = 8000
DEFAULT_MAX_ANSWER_TOKENS = 2048


@dataclass
class GcmConfig:
    """Runtime config assembled from env vars and CLI flags."""

    groq_api_key: str = ""
    model: str = DEFAULT_MODEL
    base_url: str = GROQ_BASE_URL
    token_budget: int = DEFAULT_TOKEN_BUDGET
    max_answer_tokens: int = DEFAULT_MAX_ANSWER_TOKENS
    storage_path: Optional[str] = None
    github_token: Optional[str] = None
    system_prompt: str = field(default="")

    def __post_init__(self) -> None:
        if not self.groq_api_key:
            self.groq_api_key = os.environ.get("GROQ_API_KEY", "")
        if not self.github_token:
            self.github_token = os.environ.get("GITHUB_TOKEN") or os.environ.get("GH_TOKEN")
        if not self.system_prompt:
            self.system_prompt = (
                "You are a senior software engineer answering questions about a codebase. "
                "Use ONLY the provided code context to answer. "
                "Cite file paths and symbol names when relevant. "
                "If the context is insufficient, say so — do not guess."
            )

    def validate(self) -> Optional[str]:
        """Return an error message if config is invalid, else None."""
        if not self.groq_api_key:
            return (
                "GROQ_API_KEY not set. Get a free key at https://console.groq.com "
                "and export GROQ_API_KEY=gsk_..."
            )
        return None
