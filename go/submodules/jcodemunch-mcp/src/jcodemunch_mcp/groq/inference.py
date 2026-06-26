"""Groq inference — streaming chat completions via OpenAI-compatible API."""

import sys
from typing import Iterator, Optional

from .config import GcmConfig


def _get_client(cfg: GcmConfig):
    """Lazily construct an OpenAI client pointed at Groq."""
    try:
        from openai import OpenAI
    except ImportError:
        print(
            "Error: openai package required. Install with:\n"
            "  pip install jcodemunch-mcp[groq]",
            file=sys.stderr,
        )
        sys.exit(1)

    return OpenAI(api_key=cfg.groq_api_key, base_url=cfg.base_url)


def ask(
    cfg: GcmConfig,
    context: str,
    question: str,
    history: Optional[list[dict]] = None,
) -> str:
    """Send question + context to Groq and return the full response (non-streaming)."""
    client = _get_client(cfg)

    messages = [{"role": "system", "content": cfg.system_prompt}]

    if history:
        messages.extend(history)

    user_content = f"## Code Context\n\n{context}\n\n## Question\n\n{question}" if context else question
    messages.append({"role": "user", "content": user_content})

    response = client.chat.completions.create(
        model=cfg.model,
        messages=messages,
        max_tokens=cfg.max_answer_tokens,
        temperature=0.1,
    )

    return response.choices[0].message.content or ""


def ask_stream(
    cfg: GcmConfig,
    context: str,
    question: str,
    history: Optional[list[dict]] = None,
) -> Iterator[str]:
    """Send question + context to Groq and yield response tokens as they arrive."""
    client = _get_client(cfg)

    messages = [{"role": "system", "content": cfg.system_prompt}]

    if history:
        messages.extend(history)

    user_content = f"## Code Context\n\n{context}\n\n## Question\n\n{question}" if context else question
    messages.append({"role": "user", "content": user_content})

    stream = client.chat.completions.create(
        model=cfg.model,
        messages=messages,
        max_tokens=cfg.max_answer_tokens,
        temperature=0.1,
        stream=True,
    )

    for chunk in stream:
        delta = chunk.choices[0].delta if chunk.choices else None
        if delta and delta.content:
            yield delta.content
