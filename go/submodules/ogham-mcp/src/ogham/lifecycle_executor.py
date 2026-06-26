"""Background executor for lifecycle side-effects.

open_editing_window and strengthen_edges are fire-and-forget from the
search hot path. They run on a small thread pool so search returns
without waiting for their DB round-trips.

Failure is logged, never propagated. Tests call flush() to synchronize
before asserting on post-search state.
"""

from __future__ import annotations

import atexit
import logging
import os
from concurrent.futures import Future, ThreadPoolExecutor
from typing import Any, Callable

logger = logging.getLogger(__name__)

_MAX_WORKERS = int(os.environ.get("OGHAM_LIFECYCLE_WORKERS", "4"))

_executor: ThreadPoolExecutor | None = None
_futures: list[Future] = []


def _get_executor() -> ThreadPoolExecutor:
    global _executor
    if _executor is None:
        _executor = ThreadPoolExecutor(
            max_workers=_MAX_WORKERS,
            thread_name_prefix="ogham-lifecycle",
        )
    return _executor


def submit(fn: Callable[..., Any], *args: Any, **kwargs: Any) -> Future:
    """Schedule fn(*args, **kwargs) on the background pool. Exceptions are logged, not raised."""

    def _wrapped() -> None:
        try:
            fn(*args, **kwargs)
        except Exception:
            logger.warning("lifecycle background task %s failed", fn.__name__, exc_info=True)

    future = _get_executor().submit(_wrapped)
    _futures.append(future)
    return future


def flush(timeout: float = 10.0) -> int:
    """Block until all pending tasks complete. Returns number flushed.

    Only used by tests and at process exit. Production code never needs this.
    """
    global _futures
    pending = [f for f in _futures if not f.done()]
    for f in pending:
        try:
            f.result(timeout=timeout)
        except Exception:
            logger.warning("lifecycle flush task failed", exc_info=True)
    _futures = [f for f in _futures if not f.done()]
    return len(pending)


def _shutdown() -> None:
    flush(timeout=5.0)
    global _executor
    if _executor is not None:
        _executor.shutdown(wait=True, cancel_futures=False)
        _executor = None


atexit.register(_shutdown)
