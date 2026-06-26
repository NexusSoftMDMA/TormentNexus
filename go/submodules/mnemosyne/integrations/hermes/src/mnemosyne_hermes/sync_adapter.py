"""
Sync adapter for the standalone mnemosyne-hermes plugin.

Same API as hermes_memory_provider.sync_adapter.SyncAdapter but uses
standard pip-installed imports (mnemosyne-memory core, not path-hacked).
"""

from __future__ import annotations

import json
import logging
import os
import threading
import urllib.request
import urllib.error
from pathlib import Path
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


class SyncAdapter:
    """Standalone sync adapter wrapping Mnemosyne's SyncEngine."""

    def __init__(self, beam_instance=None, config: Optional[Dict[str, Any]] = None):
        self._beam = beam_instance
        self._config = config or {}
        self._engine: Any = None
        self._error: Optional[str] = None

        self.remote = self._string("remote", "")
        self.encrypt_enabled = self._resolve_bool("encrypt", False)
        self.encryption_key = self._string("key", "")
        self.auth_token = self._string("token", "")
        self.mode = self._string("mode", "bidirectional")

        self._build_engine()

    def _string(self, key: str, default: str = "") -> str:
        env_key = f"MNEMOSYNE_SYNC_{key.upper()}"
        env_val = os.environ.get(env_key, "").strip()
        if env_val:
            return env_val
        return str(self._config.get(key, default)).strip()

    def _resolve_bool(self, key: str, default: bool = False) -> bool:
        val = self._string(key, str(default)).lower()
        return val in ("1", "true", "yes", "on")

    def _build_engine(self) -> None:
        try:
            from mnemosyne.core.sync import SyncEngine, SyncEncryption
            encryption = None
            if self.encrypt_enabled and self.encryption_key:
                encryption = SyncEncryption(key=self.encryption_key)

            # Use a BeamMemory bound to the default DB
            from mnemosyne.core.beam import BeamMemory
            beam = BeamMemory(session_id="sync-adapter")
            self._engine = SyncEngine(beam_instance=beam, encryption=encryption)
        except Exception as exc:
            self._error = str(exc)

    @property
    def is_ready(self) -> bool:
        return self._engine is not None

    def handle_tool_call(self, tool_name: str, args: dict) -> str:
        if not self.is_ready:
            return json.dumps({
                "status": "error",
                "error": f"Sync unavailable: {self._error or 'not initialized'}",
            })

        try:
            if tool_name == "mnemosyne_sync_push":
                return self._push()
            elif tool_name == "mnemosyne_sync_pull":
                return self._pull()
            elif tool_name == "mnemosyne_sync_status":
                return self._status()
        except Exception as exc:
            return json.dumps({"status": "error", "error": str(exc)})
        return json.dumps({"status": "error", "error": f"Unknown: {tool_name}"})

    def _push(self) -> str:
        if not self.remote:
            return json.dumps({"status": "error", "error": "No remote configured."})
        cursor = self._engine._meta_get("last_sync_cursor") or ""
        changes = self._engine.pull_changes(since_cursor=cursor or None, limit=500)
        events = changes.get("events", [])
        if not events:
            return json.dumps({"status": "ok", "pushed": 0})
        result = self._post("/sync/push", {"events": events})
        accepted = result.get("accepted", 0)
        cursor = result.get("next_cursor") or changes.get("next_cursor", "")
        if cursor:
            self._engine._meta_set("last_sync_cursor", cursor)
        return json.dumps({"status": "ok", "pushed": accepted})

    def _pull(self) -> str:
        if not self.remote:
            return json.dumps({"status": "error", "error": "No remote configured."})
        cursor = self._engine._meta_get("last_sync_cursor") or ""
        result = self._post("/sync/pull", {"since_token": cursor or None})
        incoming = result.get("events", [])
        if not incoming:
            return json.dumps({"status": "ok", "pulled": 0})
        push_result = self._engine.push_changes(incoming)
        cursor = result.get("next_cursor", "")
        if cursor:
            self._engine._meta_set("last_sync_cursor", cursor)
        return json.dumps({"status": "ok", "pulled": push_result.get("accepted", 0)})

    def _status(self) -> str:
        cursor = self._engine._meta_get("last_sync_cursor") or ""
        device = getattr(self._engine, "device_id", "unknown")
        try:
            row = self._engine.conn.execute("SELECT COUNT(*) FROM memory_events").fetchone()
            count = row[0] if row else 0
        except Exception:
            count = 0
        return json.dumps({
            "status": "ok",
            "device_id": device,
            "remote": self.remote or "(unconfigured)",
            "encryption": "enabled" if self.encrypt_enabled else "disabled",
            "mode": self.mode,
            "local_events": count,
            "last_cursor": cursor[:30] + "..." if len(cursor) > 30 else (cursor or "none"),
        })

    def _post(self, path: str, payload: dict) -> dict:
        url = self.remote.rstrip("/") + path
        data = json.dumps(payload).encode("utf-8")
        headers = {
            "Content-Type": "application/json",
            "User-Agent": "mnemosyne-hermes-sync/0.2.0",
        }
        if self.auth_token:
            headers["Authorization"] = f"Bearer {self.auth_token}"
        req = urllib.request.Request(url, data=data, headers=headers, method="POST")
        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                return json.loads(resp.read().decode("utf-8"))
        except Exception as exc:
            return {"status": "error", "error": str(exc)}
