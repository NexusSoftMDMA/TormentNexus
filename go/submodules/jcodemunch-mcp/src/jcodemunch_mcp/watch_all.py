"""watch-all — auto-discover every locally-indexed repo and keep it fresh.

The design deliberately stays inside the existing `WatcherManager` abstraction
instead of shelling out per-event. Discovery is registry-driven (read the
same SQLite index files jcodemunch already maintains) rather than
polling `~/.code-index/*.db` from outside.

Public surface:

    discover_local_repos(storage_path=None) -> list[str]
        Returns absolute, existing `source_root` paths for every local index.

    async watch_all(...)
        Long-running coroutine that watches every discovered repo, rediscovers
        on an interval, and shuts down cleanly on SIGINT/SIGTERM.
"""
from __future__ import annotations

import asyncio
import logging
import os
import signal
import sys
from pathlib import Path
from typing import IO, Optional

from .storage import IndexStore
from .watcher import DEFAULT_DEBOUNCE_MS, WatcherManager

logger = logging.getLogger(__name__)

DEFAULT_REDISCOVER_INTERVAL_S = 30.0


def discover_local_repos(storage_path: Optional[str] = None) -> list[str]:
    """Return resolved on-disk paths for every locally-indexed repo.

    GitHub repos (empty `source_root`) and indexes whose source_root no longer
    exists on disk are skipped — the latter protects against watchdog blowing
    up when a repo was deleted out from under the index.
    """
    store = IndexStore(base_path=storage_path) if storage_path else IndexStore()
    repos: list[str] = []
    for entry in store.list_repos():
        src = (entry.get("source_root") or "").strip()
        if not src:
            continue
        path = Path(src).expanduser()
        try:
            if not path.is_dir():
                continue
            repos.append(str(path.resolve()))
        except OSError:
            logger.debug("Unreachable source_root: %s", src, exc_info=True)
    return sorted(set(repos))


def _install_signal_handlers(loop: asyncio.AbstractEventLoop, stop: asyncio.Event) -> None:
    def _request_stop() -> None:
        if not stop.is_set():
            stop.set()

    if sys.platform == "win32":
        # add_signal_handler is not supported on Windows ProactorEventLoop.
        # KeyboardInterrupt from Ctrl-C surfaces as CancelledError in asyncio.run().
        return
    for sig in (signal.SIGINT, signal.SIGTERM):
        try:
            loop.add_signal_handler(sig, _request_stop)
        except (NotImplementedError, RuntimeError):
            logger.debug("Could not install handler for %s", sig, exc_info=True)


async def watch_all(
    debounce_ms: int = DEFAULT_DEBOUNCE_MS,
    use_ai_summaries: bool = True,
    storage_path: Optional[str] = None,
    extra_ignore_patterns: Optional[list[str]] = None,
    follow_symlinks: bool = False,
    rediscover_interval_s: float = DEFAULT_REDISCOVER_INTERVAL_S,
    quiet: bool = False,
    log_file_handle: Optional[IO] = None,
) -> None:
    """Watch every locally-indexed repo; rediscover on an interval.

    Repos added to the registry while running are picked up on the next
    rediscovery pass. Repos whose source_root disappears are dropped.
    """
    manager = WatcherManager(
        debounce_ms=debounce_ms,
        use_ai_summaries=use_ai_summaries,
        storage_path=storage_path,
        extra_ignore_patterns=extra_ignore_patterns,
        follow_symlinks=follow_symlinks,
        quiet=quiet,
        log_file_handle=log_file_handle,
    )

    stop_event = asyncio.Event()
    try:
        loop = asyncio.get_running_loop()
        _install_signal_handlers(loop, stop_event)
    except RuntimeError:
        pass

    async def _rediscover_loop() -> None:
        while not stop_event.is_set():
            try:
                discovered = set(discover_local_repos(storage_path))
                watched = set(manager.list_folders())
                for folder in sorted(discovered - watched):
                    await manager.add_folder(folder)
                for folder in sorted(watched - discovered):
                    await manager.remove_folder(folder)
            except Exception:
                logger.warning("rediscover pass failed", exc_info=True)
            try:
                await asyncio.wait_for(stop_event.wait(), timeout=rediscover_interval_s)
            except asyncio.TimeoutError:
                continue

    # Seed discovery before run() starts so initial repos are watched
    # by the time the caller sees log lines.
    initial = discover_local_repos(storage_path)
    for folder in initial:
        await manager.add_folder(folder)

    rediscover_task = asyncio.create_task(_rediscover_loop(), name="watch-all:rediscover")
    run_task = asyncio.create_task(manager.run(), name="watch-all:manager")

    try:
        done, pending = await asyncio.wait(
            {rediscover_task, run_task, asyncio.create_task(stop_event.wait())},
            return_when=asyncio.FIRST_COMPLETED,
        )
        for t in pending:
            t.cancel()
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass
    finally:
        stop_event.set()
        manager.stop()
        for folder in list(manager.list_folders()):
            try:
                await manager.remove_folder(folder)
            except Exception:
                logger.debug("remove_folder failed on shutdown: %s", folder, exc_info=True)
        for t in (rediscover_task, run_task):
            t.cancel()
            try:
                await asyncio.wait_for(t, timeout=2.0)
            except (asyncio.TimeoutError, asyncio.CancelledError, Exception):
                pass


def storage_path_default() -> str:
    return os.environ.get("CODE_INDEX_PATH") or str(Path.home() / ".code-index")
