"""Turn budget tracking for auto-compact mode."""

import threading
import time
from typing import Optional


class TurnBudget:
    """Track token usage across tool calls in a turn.

    A "turn" is defined as a series of tool calls with gaps less than
    turn_gap_seconds. When the gap exceeds this threshold, a new turn
    starts and counters reset. The reset fires in EVERY reader (jcm#327):
    readers that run during tool execution (plan_turn's budget advisor,
    should_compact) would otherwise report the previous turn's exhausted
    counter after an idle gap, because record_output — the only reset
    site before the fix — runs post-dispatch.

    Scope note: this is a process-global singleton, so the budget spans
    every repo served by one server process. The wall-clock gap is a
    heuristic for conversational turn boundaries; during continuous agent
    use (calls spaced under turn_gap_seconds) it behaves as a rolling
    exploration budget rather than a strict per-turn one.
    """

    def __init__(
        self,
        turn_budget_tokens: int = 20000,
        turn_gap_seconds: float = 30.0,
    ):
        self._lock = threading.Lock()
        self._turn_budget = turn_budget_tokens
        self._turn_gap = turn_gap_seconds
        self._last_call_ts: float = 0.0
        self._turn_tokens: int = 0
        self._turn_call_count: int = 0

    def is_new_turn(self) -> bool:
        """Check if enough time has passed for a new turn."""
        return time.time() - self._last_call_ts > self._turn_gap

    def _maybe_reset_locked(self, now: float) -> None:
        """Zero the counters when the idle gap marks a new turn.

        Caller must hold self._lock. Does NOT advance _last_call_ts —
        only record_output marks activity, so back-to-back readers after
        an idle gap each still see the fresh turn (jcm#327).
        """
        if now - self._last_call_ts > self._turn_gap:
            self._turn_tokens = 0
            self._turn_call_count = 0

    def record_output(self, token_count: int) -> dict:
        """Record output tokens and return budget info.

        Returns:
            Dict with turn_tokens_used, turn_budget_remaining, and
            budget_warning (if over 80% or exhausted).
            Empty dict if budget is 0 (disabled).
        """
        if self._turn_budget == 0:
            return {}

        with self._lock:
            now = time.time()
            self._maybe_reset_locked(now)
            self._last_call_ts = now
            self._turn_tokens += token_count
            self._turn_call_count += 1

            remaining = max(0, self._turn_budget - self._turn_tokens)
            used_percent = (self._turn_tokens / self._turn_budget) * 100

            result = {
                "turn_tokens_used": self._turn_tokens,
                "turn_budget_remaining": remaining,
            }

            if used_percent >= 100:
                result["budget_warning"] = (
                    f"Turn budget exhausted ({self._turn_tokens}/{self._turn_budget} tokens). "
                    "Consider stopping exploration and working with current context."
                )
            elif used_percent >= 80:
                result["budget_warning"] = (
                    f"Turn budget {used_percent:.0f}% used ({self._turn_tokens}/{self._turn_budget} tokens). "
                    "Focus remaining reads on highest-value files."
                )

            return result

    def percent_used(self) -> float:
        """Return fraction of turn budget used (0.0-1.0+). 0.0 if disabled."""
        if self._turn_budget == 0:
            return 0.0
        with self._lock:
            self._maybe_reset_locked(time.time())
            return self._turn_tokens / self._turn_budget

    def is_enabled(self) -> bool:
        """Return True if turn budget tracking is active."""
        return self._turn_budget > 0

    def configure(self, budget_tokens: int, gap_seconds: float) -> None:
        """Reconfigure budget parameters (thread-safe)."""
        with self._lock:
            self._turn_budget = budget_tokens
            self._turn_gap = gap_seconds

    def should_compact(self) -> bool:
        """Return True if results should be auto-compacted due to budget pressure."""
        if self._turn_budget == 0:
            return False
        with self._lock:
            self._maybe_reset_locked(time.time())
            return self._turn_tokens >= self._turn_budget * 0.8

    def reset(self) -> None:
        """Manually reset the turn counters."""
        with self._lock:
            self._turn_tokens = 0
            self._turn_call_count = 0
            self._last_call_ts = time.time()


# Singleton instance
_budget: Optional[TurnBudget] = None
_budget_lock = threading.Lock()


def get_turn_budget() -> TurnBudget:
    """Get the singleton TurnBudget instance."""
    global _budget
    with _budget_lock:
        if _budget is None:
            _budget = TurnBudget()
        return _budget