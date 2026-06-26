"""
Multi-backend LLM Pool: LM Studio (local) -> OpenRouter (free cloud fallback)
OpenAI-compatible API for both backends.
"""
import os
import re
import json
import time
import logging
import requests


def stringify_field(value):
    if value is None:
        return ""
    if isinstance(value, str):
        return value
    if isinstance(value, (list, tuple, set)):
        parts = [str(item).strip() for item in value if str(item).strip()]
        return ", ".join(parts)
    return str(value)


class LLMPool:
    # LM Studio local models (ordered by speed - smaller models first for batch processing)
    LMSTUDIO_MODELS = [
        "liquid/lfm2.5-1.2b",
        "gemma-4-e2b-uncensored-hauhaucs-aggressive",
        "gemma-4-e4b-uncensored-hauhaucs-aggressive",
        "gemma-4-26b-a4b-it-heretic-ara",
        "gemma-4-31b-it-heretic",
        "gemma-4-31b-it-abliterated",
    ]

    # OpenRouter free models (ordered by capability)
    OPENROUTER_FREE_MODELS = [
        "google/gemma-4-31b-it:free",
        "google/gemma-4-26b-a4b-it:free",
        "nvidia/nemotron-3-super-120b-a12b:free",
        "minimax/minimax-m2.5:free",
        "qwen/qwen3-next-80b-a3b-instruct:free",
        "nvidia/nemotron-3-nano-30b-a3b:free",
        "openrouter/free",
    ]

    def __init__(self, logger=None, sleep_seconds=30):
        self.logger = logger
        self.sleep_seconds = sleep_seconds
        self.lmstudio_url = os.environ.get("LMSTUDIO_URL", "http://localhost:1234")
        self.openrouter_key = os.environ.get("OPENROUTER_API_KEY", "")
        self.openrouter_url = "https://openrouter.ai/api/v1/chat/completions"

        self.active_model_name = None
        self.active_backend = None  # "lmstudio" or "openrouter"
        self.last_error_summary = None
        self.last_backoff_seconds = None

        # Discover which LM Studio models are actually loaded
        self.lmstudio_available = self._discover_lmstudio()
        self.all_backends = self._build_backend_list()

        self._log("info", f"LLM Pool initialized: {len(self.lmstudio_available)} LM Studio models, "
                         f"{len(self.OPENROUTER_FREE_MODELS)} OpenRouter free models")

    def _log(self, level, message):
        if self.logger and hasattr(self.logger, level):
            getattr(self.logger, level)(message)

    def _discover_lmstudio(self):
        """Query LM Studio for currently loaded models."""
        try:
            resp = requests.get(f"{self.lmstudio_url}/v1/models", timeout=5)
            if resp.status_code == 200:
                data = resp.json()
                loaded = [m["id"] for m in data.get("data", [])]
                # Filter to only non-embedding models we know about, preserving quality order
                available = [m for m in self.LMSTUDIO_MODELS if m in loaded]
                if not available:
                    # Use any non-embedding model that's loaded
                    available = [m for m in loaded if "embedding" not in m.lower()]
                self._log("info", f"LM Studio models discovered: {available}")
                return available
        except Exception as e:
            self._log("warning", f"LM Studio not available: {e}")
        return []

    def _build_backend_list(self):
        """Build ordered list of (backend, model) tuples."""
        backends = []
        for m in self.lmstudio_available:
            backends.append(("lmstudio", m))
        for m in self.OPENROUTER_FREE_MODELS:
            backends.append(("openrouter", m))
        return backends

    def _call_lmstudio(self, model, messages):
        """Call LM Studio via OpenAI-compatible API."""
        resp = requests.post(
            f"{self.lmstudio_url}/v1/chat/completions",
            json={
                "model": model,
                "messages": messages,
                "temperature": 0.3,
                "max_tokens": 2048,
            },
            timeout=60,
        )
        resp.raise_for_status()
        data = resp.json()
        return data["choices"][0]["message"]["content"]

    def _call_openrouter(self, model, messages):
        """Call OpenRouter via OpenAI-compatible API."""
        resp = requests.post(
            self.openrouter_url,
            headers={
                "Authorization": f"Bearer {self.openrouter_key}",
                "Content-Type": "application/json",
                "HTTP-Referer": "https://github.com/robertpelloni/bobbybookmarks",
            },
            json={
                "model": model,
                "messages": messages,
                "temperature": 0.3,
                "max_tokens": 2048,
            },
            timeout=60,
        )
        resp.raise_for_status()
        data = resp.json()
        return data["choices"][0]["message"]["content"]

    def _summarize_error(self, error):
        return " ".join(str(error).split())[:240]

    def generate_content(self, prompt, context_label=""):
        """
        Try all backends in order: LM Studio models first, then OpenRouter free models.
        Returns (response_text, model_name) or (None, None) on failure.
        """
        last_error = None

        for backend, model in self.all_backends:
            try:
                messages = [{"role": "user", "content": prompt}]

                if backend == "lmstudio":
                    text = self._call_lmstudio(model, messages)
                else:
                    text = self._call_openrouter(model, messages)

                self.active_model_name = model
                self.active_backend = backend
                self.last_error_summary = None
                self.last_backoff_seconds = None
                backend_tag = "LM Studio" if backend == "lmstudio" else "OpenRouter"
                self._log("info", f"[{backend_tag}] {model} succeeded for {context_label}")
                return text, f"{backend_tag}/{model}"

            except Exception as e:
                last_error = e
                msg = self._summarize_error(e)
                backend_tag = "LM Studio" if backend == "lmstudio" else "OpenRouter"
                self._log("warning", f"[{backend_tag}] {model} failed for {context_label}: {msg}")
                continue

        # All backends failed
        if last_error:
            self.last_error_summary = self._summarize_error(last_error)
            self._log("error", f"All LLM backends failed while {context_label}. Sleeping {self.sleep_seconds}s.")
        self.last_backoff_seconds = self.sleep_seconds
        time.sleep(self.sleep_seconds)
        return None, None

    def refresh_lmstudio(self):
        """Re-discover LM Studio models (in case user loads new ones)."""
        self.lmstudio_available = self._discover_lmstudio()
        self.all_backends = self._build_backend_list()
