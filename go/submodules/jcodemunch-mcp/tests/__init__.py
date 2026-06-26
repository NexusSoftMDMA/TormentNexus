"""Test package."""

import sys
from pathlib import Path


def _platform_path(unix_path: str) -> Path:
    """Convert Unix-style path to platform-appropriate path for testing.

    On Unix: returns Path(unix_path) unchanged.
    On Windows: converts "/work" to "C:/work" to ensure is_absolute() is True.
    """
    if sys.platform == "win32":
        if unix_path.startswith("/"):
            return Path("C:" + unix_path.replace("/", "/"))
    return Path(unix_path)


def _platform_path_str(unix_path: str) -> str:
    """Convert Unix-style path to platform-appropriate path string for config files.

    Returns forward-slash paths on Windows so the result is safe to embed in JSON.
    Python's pathlib accepts forward slashes on Windows, so this is valid at runtime.
    """
    return _platform_path(unix_path).as_posix()
