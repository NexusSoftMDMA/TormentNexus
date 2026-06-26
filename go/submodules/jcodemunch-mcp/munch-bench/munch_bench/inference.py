"""Inference layer — calls LLM providers to answer codebase questions."""

from __future__ import annotations

import time
from dataclasses import dataclass
from typing import Optional


SYSTEM_PROMPT = (
    "You are a senior software engineer answering questions about a codebase. "
    "Use the provided code context to give accurate, concise answers. "
    "If the context doesn't contain enough information, say so."
)


@dataclass
class InferenceResult:
    """Result of a single inference call."""

    answer: str
    model: str
    provider: str
    wall_time_s: float
    input_tokens: int
    output_tokens: int
    cost_usd: float


# Approximate costs per 1M tokens (input, output)
COST_TABLE: dict[str, tuple[float, float]] = {
    # Groq
    "llama-3.3-70b-versatile": (0.59, 0.79),
    "llama-3.1-8b-instant": (0.05, 0.08),
    "llama3-70b-8192": (0.59, 0.79),
    "llama3-8b-8192": (0.05, 0.08),
    "gemma2-9b-it": (0.20, 0.20),
    "mixtral-8x7b-32768": (0.24, 0.24),
    "deepseek-r1-distill-llama-70b": (0.75, 0.99),
    # OpenAI
    "gpt-4o": (2.50, 10.00),
    "gpt-4o-mini": (0.15, 0.60),
    "gpt-4.1": (2.00, 8.00),
    "gpt-4.1-mini": (0.40, 1.60),
    "gpt-4.1-nano": (0.10, 0.40),
    # Anthropic
    "claude-sonnet-4-6": (3.00, 15.00),
    "claude-haiku-4-5-20251001": (0.80, 4.00),
    "claude-opus-4-6": (15.00, 75.00),
}


def _estimate_cost(model: str, input_tokens: int, output_tokens: int) -> float:
    """Estimate cost in USD from token counts."""
    costs = COST_TABLE.get(model, (1.0, 3.0))  # conservative default
    return (input_tokens * costs[0] + output_tokens * costs[1]) / 1_000_000


def infer_groq(
    context: str,
    question: str,
    model: str = "llama-3.3-70b-versatile",
    api_key: Optional[str] = None,
) -> InferenceResult:
    """Call Groq API (OpenAI-compatible) for inference."""
    import openai
    import os

    client = openai.OpenAI(
        api_key=api_key or os.environ.get("GROQ_API_KEY", ""),
        base_url="https://api.groq.com/openai/v1",
    )

    t0 = time.perf_counter()
    response = client.chat.completions.create(
        model=model,
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": f"## Code Context\n\n{context}\n\n## Question\n\n{question}"},
        ],
        temperature=0.0,
        max_tokens=2048,
    )
    elapsed = time.perf_counter() - t0

    usage = response.usage
    input_tok = usage.prompt_tokens if usage else 0
    output_tok = usage.completion_tokens if usage else 0

    return InferenceResult(
        answer=response.choices[0].message.content or "",
        model=model,
        provider="groq",
        wall_time_s=elapsed,
        input_tokens=input_tok,
        output_tokens=output_tok,
        cost_usd=_estimate_cost(model, input_tok, output_tok),
    )


def infer_openai(
    context: str,
    question: str,
    model: str = "gpt-4o-mini",
    api_key: Optional[str] = None,
) -> InferenceResult:
    """Call OpenAI API for inference."""
    import openai
    import os

    client = openai.OpenAI(api_key=api_key or os.environ.get("OPENAI_API_KEY", ""))

    t0 = time.perf_counter()
    response = client.chat.completions.create(
        model=model,
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": f"## Code Context\n\n{context}\n\n## Question\n\n{question}"},
        ],
        temperature=0.0,
        max_tokens=2048,
    )
    elapsed = time.perf_counter() - t0

    usage = response.usage
    input_tok = usage.prompt_tokens if usage else 0
    output_tok = usage.completion_tokens if usage else 0

    return InferenceResult(
        answer=response.choices[0].message.content or "",
        model=model,
        provider="openai",
        wall_time_s=elapsed,
        input_tokens=input_tok,
        output_tokens=output_tok,
        cost_usd=_estimate_cost(model, input_tok, output_tok),
    )


def infer_anthropic(
    context: str,
    question: str,
    model: str = "claude-sonnet-4-6",
    api_key: Optional[str] = None,
) -> InferenceResult:
    """Call Anthropic API for inference."""
    import anthropic
    import os

    client = anthropic.Anthropic(api_key=api_key or os.environ.get("ANTHROPIC_API_KEY", ""))

    t0 = time.perf_counter()
    response = client.messages.create(
        model=model,
        system=SYSTEM_PROMPT,
        messages=[
            {"role": "user", "content": f"## Code Context\n\n{context}\n\n## Question\n\n{question}"},
        ],
        temperature=0.0,
        max_tokens=2048,
    )
    elapsed = time.perf_counter() - t0

    input_tok = response.usage.input_tokens
    output_tok = response.usage.output_tokens
    answer = response.content[0].text if response.content else ""

    return InferenceResult(
        answer=answer,
        model=model,
        provider="anthropic",
        wall_time_s=elapsed,
        input_tokens=input_tok,
        output_tokens=output_tok,
        cost_usd=_estimate_cost(model, input_tok, output_tok),
    )


PROVIDER_MAP = {
    "groq": infer_groq,
    "openai": infer_openai,
    "anthropic": infer_anthropic,
}


def infer(
    context: str,
    question: str,
    provider: str = "groq",
    model: Optional[str] = None,
    api_key: Optional[str] = None,
) -> InferenceResult:
    """Dispatch inference to the appropriate provider."""
    defaults = {
        "groq": "llama-3.3-70b-versatile",
        "openai": "gpt-4o-mini",
        "anthropic": "claude-sonnet-4-6",
    }
    fn = PROVIDER_MAP.get(provider)
    if fn is None:
        raise ValueError(f"Unknown provider: {provider}. Choose from: {', '.join(PROVIDER_MAP)}")
    model = model or defaults.get(provider, "")
    return fn(context, question, model=model, api_key=api_key)
