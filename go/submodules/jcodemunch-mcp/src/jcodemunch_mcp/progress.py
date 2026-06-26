"""MCP progress notification helper for long-running tools.

Emits ``notifications/progress`` so MCP hosts (e.g. VS Code) can show
a live inline indicator.  Zero token cost — notifications go to the host,
never the model.  No-op when the client omits ``progressToken``.
"""

from __future__ import annotations

import asyncio
import logging
import threading
from typing import Callable, Optional

logger = logging.getLogger(__name__)

# Type alias: (progress, total, message) → None
ProgressNotify = Callable[[float, Optional[float], Optional[str]], None]

# Type alias for tool-level callbacks: (done, total, detail) → None
ProgressHook = Optional[Callable[[int, int, str], None]]


class ProgressReporter:
    """Emit monotonic MCP progress notifications.

    Thread-safe: ``update()`` and ``finish()`` may be called from worker
    threads (e.g. inside ``asyncio.to_thread``).

    No fake drift, no pulse threads — progress reflects real completed work.
    If a slow sub-step stalls, the bar stalls.  That's honest.
    """

    __slots__ = (
        "_notify", "_label", "_bar_width", "_lock",
        "_last_sent", "_done", "_total", "_finished",
    )

    def __init__(
        self,
        notify: Optional[ProgressNotify],
        label: str,
        *,
        bar_width: int = 12,
    ) -> None:
        self._notify = notify
        self._label = label
        self._bar_width = bar_width
        self._lock = threading.Lock()
        self._last_sent: float = 0.0
        self._done: int = 0
        self._total: int = 0
        self._finished: bool = False

    def start(self, total: int = 0, detail: str = "Starting") -> None:
        """Emit initial 0% notification."""
        if self._notify is None:
            return
        with self._lock:
            self._total = max(total, 0)
        self._send(0.0, detail)

    def update(self, done: int, total: int, detail: str = "") -> None:
        """Emit progress for real completed work."""
        if self._notify is None:
            return
        total = max(int(total), 1)
        done = max(0, min(int(done), total))
        with self._lock:
            if self._finished:
                return
            self._done = done
            self._total = total
            progress = done / total
            # Monotonic: never go backwards
            if progress <= self._last_sent:
                return
        self._send(progress, detail)

    def finish(self, detail: str = "Complete") -> None:
        """Emit 100% notification."""
        if self._notify is None:
            return
        with self._lock:
            if self._finished:
                return
            self._finished = True
            if self._total > 0:
                self._done = self._total
        self._send(1.0, detail)

    def _send(self, progress: float, detail: str) -> None:
        with self._lock:
            progress = max(progress, self._last_sent)
            progress = min(progress, 1.0)
            self._last_sent = progress
            message = self._format(progress, detail)
            try:
                self._notify(progress, 1.0, message)
            except Exception:
                logger.debug("progress notification failed", exc_info=True)

    def _format(self, progress: float, detail: str) -> str:
        filled = int(progress * self._bar_width)
        bar = "[" + "#" * filled + "-" * (self._bar_width - filled) + "]"
        pct = f"{progress * 100:5.1f}%"
        parts = [self._label, bar, pct]
        if self._total > 0:
            parts.append(f"{min(self._done, self._total)}/{self._total}")
        if detail:
            parts.append(detail)
        return " ".join(parts)


def make_progress_notify(server_obj) -> Optional[ProgressNotify]:
    """Create a thread-safe MCP progress notifier for the current request.

    Returns None if the client didn't send a progressToken.
    """
    try:
        ctx = server_obj.request_context
    except LookupError:
        return None

    if ctx.meta is None or ctx.meta.progressToken is None:
        return None

    loop = asyncio.get_running_loop()
    session = ctx.session
    progress_token = ctx.meta.progressToken

    def _notify(progress: float, total: float | None, message: str | None) -> None:
        async def _send() -> None:
            try:
                await session.send_progress_notification(
                    progress_token=progress_token,
                    progress=progress,
                    total=total,
                    message=message,
                )
            except Exception:
                logger.debug("progress notification send failed", exc_info=True)

        try:
            asyncio.run_coroutine_threadsafe(_send(), loop)
        except RuntimeError:
            pass  # event loop closed or unavailable

    return _notify
