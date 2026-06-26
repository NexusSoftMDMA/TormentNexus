"""Cross-platform login-service installer for `jcodemunch-mcp watch-all`.

- Linux: systemd --user unit at ~/.config/systemd/user/jcodemunch-watch.service
- macOS: launchd plist at ~/Library/LaunchAgents/us.gravelle.jcodemunch-watch.plist
- Windows: Task Scheduler task named `jcodemunch-watch`

The installer deliberately invokes the *same interpreter* currently running
(via `sys.executable -m jcodemunch_mcp watch-all`) so the service picks up
whatever virtualenv the user installed into. This avoids the `uvx` round-trip
that jcrefresher pays per-event and removes a whole class of PATH issues.
"""
from __future__ import annotations

import logging
import os
import platform
import shutil
import subprocess
import sys
from pathlib import Path

logger = logging.getLogger(__name__)

SERVICE_NAME = "jcodemunch-watch"
LAUNCHD_LABEL = "us.gravelle.jcodemunch-watch"


class InstallerError(RuntimeError):
    pass


# ── Path helpers ────────────────────────────────────────────────────────────


def _systemd_unit_path() -> Path:
    return Path.home() / ".config" / "systemd" / "user" / f"{SERVICE_NAME}.service"


def _launchd_plist_path() -> Path:
    return Path.home() / "Library" / "LaunchAgents" / f"{LAUNCHD_LABEL}.plist"


def _log_dir() -> Path:
    base = Path(os.environ.get("CODE_INDEX_PATH") or (Path.home() / ".code-index"))
    return base / "logs"


def _exec_cmd() -> list[str]:
    """How the service should invoke the watcher."""
    return [sys.executable, "-m", "jcodemunch_mcp", "watch-all"]


# ── systemd (Linux) ─────────────────────────────────────────────────────────


_SYSTEMD_TEMPLATE = """[Unit]
Description=jcodemunch-mcp: auto-reindex every locally-indexed repo
After=default.target

[Service]
Type=simple
ExecStart={exec_cmd}
Restart=on-failure
RestartSec=5
StandardOutput=append:{log_dir}/watch.log
StandardError=append:{log_dir}/watch.err
Environment=PYTHONUNBUFFERED=1
{env_lines}

[Install]
WantedBy=default.target
"""


def _systemd_env_lines() -> str:
    """Forward CODE_INDEX_PATH and JCODEMUNCH_* env into the unit."""
    lines = []
    for key, val in os.environ.items():
        if key == "CODE_INDEX_PATH" or key.startswith("JCODEMUNCH_"):
            lines.append(f"Environment={key}={val}")
    return "\n".join(lines)


def _install_systemd() -> dict:
    if shutil.which("systemctl") is None:
        raise InstallerError("systemctl not found — is this a systemd system?")
    unit_path = _systemd_unit_path()
    unit_path.parent.mkdir(parents=True, exist_ok=True)
    _log_dir().mkdir(parents=True, exist_ok=True)

    quoted = " ".join(_shell_quote(x) for x in _exec_cmd())
    unit_path.write_text(
        _SYSTEMD_TEMPLATE.format(
            exec_cmd=quoted,
            log_dir=str(_log_dir()),
            env_lines=_systemd_env_lines(),
        ),
        encoding="utf-8",
    )

    subprocess.run(["systemctl", "--user", "daemon-reload"], check=True)
    subprocess.run(["systemctl", "--user", "enable", "--now", f"{SERVICE_NAME}.service"], check=True)
    return {"platform": "systemd", "unit": str(unit_path), "status": "enabled"}


def _uninstall_systemd() -> dict:
    unit_path = _systemd_unit_path()
    if shutil.which("systemctl"):
        subprocess.run(["systemctl", "--user", "disable", "--now", f"{SERVICE_NAME}.service"], check=False)
        subprocess.run(["systemctl", "--user", "daemon-reload"], check=False)
    removed = False
    if unit_path.exists():
        unit_path.unlink()
        removed = True
    return {"platform": "systemd", "unit": str(unit_path), "removed": removed}


def _status_systemd() -> dict:
    if shutil.which("systemctl") is None:
        return {"platform": "systemd", "active": False, "reason": "systemctl not found"}
    result = subprocess.run(
        ["systemctl", "--user", "is-active", f"{SERVICE_NAME}.service"],
        capture_output=True, text=True, check=False,
    )
    state = result.stdout.strip() or result.stderr.strip()
    return {"platform": "systemd", "active": state == "active", "state": state}


# ── launchd (macOS) ─────────────────────────────────────────────────────────


_LAUNCHD_TEMPLATE = """<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>{label}</string>
  <key>ProgramArguments</key>
  <array>
{args}
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>{log_dir}/watch.log</string>
  <key>StandardErrorPath</key><string>{log_dir}/watch.err</string>
  <key>EnvironmentVariables</key>
  <dict>
{env}
  </dict>
</dict></plist>
"""


def _launchd_env_xml() -> str:
    out = []
    for key, val in os.environ.items():
        if key == "CODE_INDEX_PATH" or key == "PATH" or key.startswith("JCODEMUNCH_"):
            out.append(f"    <key>{_xml_escape(key)}</key><string>{_xml_escape(val)}</string>")
    return "\n".join(out)


def _install_launchd() -> dict:
    plist = _launchd_plist_path()
    plist.parent.mkdir(parents=True, exist_ok=True)
    _log_dir().mkdir(parents=True, exist_ok=True)
    args_xml = "\n".join(f"    <string>{_xml_escape(a)}</string>" for a in _exec_cmd())
    plist.write_text(
        _LAUNCHD_TEMPLATE.format(
            label=LAUNCHD_LABEL,
            args=args_xml,
            log_dir=str(_log_dir()),
            env=_launchd_env_xml(),
        ),
        encoding="utf-8",
    )
    subprocess.run(["launchctl", "unload", str(plist)], check=False)
    subprocess.run(["launchctl", "load", str(plist)], check=True)
    return {"platform": "launchd", "plist": str(plist), "status": "loaded"}


def _uninstall_launchd() -> dict:
    plist = _launchd_plist_path()
    subprocess.run(["launchctl", "unload", str(plist)], check=False)
    removed = False
    if plist.exists():
        plist.unlink()
        removed = True
    return {"platform": "launchd", "plist": str(plist), "removed": removed}


def _status_launchd() -> dict:
    result = subprocess.run(
        ["launchctl", "list", LAUNCHD_LABEL],
        capture_output=True, text=True, check=False,
    )
    return {"platform": "launchd", "active": result.returncode == 0, "detail": result.stdout.strip()}


# ── Task Scheduler (Windows) ────────────────────────────────────────────────


def _install_windows() -> dict:
    _log_dir().mkdir(parents=True, exist_ok=True)
    cmd_str = " ".join(_cmd_quote(x) for x in _exec_cmd())
    # schtasks /Create does not persist stdout redirection; rely on Python logging
    # (watcher.py writes to stderr) and inspect via Event Viewer if needed.
    args = [
        "schtasks", "/Create", "/F",
        "/TN", SERVICE_NAME,
        "/SC", "ONLOGON",
        "/RL", "LIMITED",
        "/TR", cmd_str,
    ]
    result = subprocess.run(args, capture_output=True, text=True, check=False)
    if result.returncode != 0:
        raise InstallerError(f"schtasks /Create failed: {result.stderr.strip() or result.stdout.strip()}")
    subprocess.run(["schtasks", "/Run", "/TN", SERVICE_NAME], check=False, capture_output=True)
    return {"platform": "schtasks", "task": SERVICE_NAME, "status": "registered"}


def _uninstall_windows() -> dict:
    result = subprocess.run(
        ["schtasks", "/Delete", "/F", "/TN", SERVICE_NAME],
        capture_output=True, text=True, check=False,
    )
    return {"platform": "schtasks", "task": SERVICE_NAME, "removed": result.returncode == 0}


def _status_windows() -> dict:
    result = subprocess.run(
        ["schtasks", "/Query", "/TN", SERVICE_NAME, "/FO", "LIST"],
        capture_output=True, text=True, check=False,
    )
    active = "Running" in result.stdout or "Ready" in result.stdout
    return {"platform": "schtasks", "active": active, "detail": result.stdout.strip()[:400]}


# ── Public dispatch ─────────────────────────────────────────────────────────


def install_service() -> dict:
    sys_ = platform.system()
    if sys_ == "Linux":
        return _install_systemd()
    if sys_ == "Darwin":
        return _install_launchd()
    if sys_ == "Windows":
        return _install_windows()
    raise InstallerError(f"Unsupported platform: {sys_}")


def uninstall_service() -> dict:
    sys_ = platform.system()
    if sys_ == "Linux":
        return _uninstall_systemd()
    if sys_ == "Darwin":
        return _uninstall_launchd()
    if sys_ == "Windows":
        return _uninstall_windows()
    raise InstallerError(f"Unsupported platform: {sys_}")


def service_status() -> dict:
    sys_ = platform.system()
    if sys_ == "Linux":
        return _status_systemd()
    if sys_ == "Darwin":
        return _status_launchd()
    if sys_ == "Windows":
        return _status_windows()
    return {"platform": sys_, "active": False, "reason": "unsupported"}


# ── escaping helpers ────────────────────────────────────────────────────────


def _shell_quote(s: str) -> str:
    if not s or any(c in s for c in ' \t"\''):
        return "'" + s.replace("'", "'\\''") + "'"
    return s


def _cmd_quote(s: str) -> str:
    if " " in s or "\t" in s:
        return '"' + s.replace('"', '\\"') + '"'
    return s


def _xml_escape(s: str) -> str:
    return (
        s.replace("&", "&amp;")
         .replace("<", "&lt;")
         .replace(">", "&gt;")
         .replace('"', "&quot;")
    )
